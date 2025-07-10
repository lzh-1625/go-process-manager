package api

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/constants"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type wsApi struct{}

var WsApi = new(wsApi)

type WsConnetInstance struct {
	WsConnect  *websocket.Conn
	wsLock     sync.Mutex
	CancelFunc context.CancelFunc
}

func (w *WsConnetInstance) Write(b []byte) {
	w.wsLock.Lock()
	defer w.wsLock.Unlock()
	w.WsConnect.WriteMessage(websocket.BinaryMessage, b)
}

func (w *WsConnetInstance) WriteString(s string) {
	w.wsLock.Lock()
	defer w.wsLock.Unlock()
	w.WsConnect.WriteMessage(websocket.TextMessage, []byte(s))
}
func (w *WsConnetInstance) Cancel() {
	w.CancelFunc()
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (w *wsApi) WebsocketHandle(ctx *gin.Context, req model.WebsocketHandleReq) (err error) {
	reqUser := getUserName(ctx)
	proc, err := logic.ProcessCtlLogic.GetProcess(req.Uuid)
	if err != nil {
		return err
	}
	if proc.HasWsConn(reqUser) {
		return errors.New("connection already exists")
	}
	if proc.Control.Controller != reqUser && !proc.VerifyControl() {
		return errors.New("insufficient permissions")
	}
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return err
	}

	log.Logger.Infow("ws连接成功")

	wsCtx, cancel := context.WithCancel(context.Background())
	wci := &WsConnetInstance{
		WsConnect:  conn,
		CancelFunc: cancel,
		wsLock:     sync.Mutex{},
	}
	proc.ReadCache(wci)
	if proc.State.State == 1 {
		proc.SetTerminalSize(req.Cols, req.Rows)
		w.startWsConnect(wci, cancel, proc, hasOprPermission(ctx, req.Uuid, constants.OPERATION_TERMINAL_WRITE))
		proc.AddConn(reqUser, wci)
		defer proc.DeleteConn(reqUser)
	}
	conn.SetCloseHandler(func(_ int, _ string) error {
		cancel()
		return nil
	})
	select {
	case <-proc.StopChan:
		log.Logger.Infow("ws连接断开", "操作类型", "进程已停止，强制断开ws连接")
	case <-time.After(time.Minute * time.Duration(config.CF.TerminalConnectTimeout)):
		log.Logger.Infow("ws连接断开", "操作类型", "连接时间超过最大时长限制")
	case <-wsCtx.Done():
		log.Logger.Infow("ws连接断开", "操作类型", "tcp连接建立已被关闭")
	}
	conn.Close()
	return
}

func (w *wsApi) WebsocketShareHandle(ctx *gin.Context, req model.WebsocketHandleReq) (err error) {
	data, err := repository.WsShare.GetWsShareDataByToken(req.Token)
	if err != nil {
		return err
	}
	if data.ExpireTime.Before(time.Now()) {
		return errors.New("share expired")
	}
	proc, err := logic.ProcessCtlLogic.GetProcess(data.Pid)
	if err != nil {
		return err
	}
	guestName := "guest-" + strconv.Itoa(int(data.ID)) // 构造访客用户名
	if proc.HasWsConn(guestName) {
		return errors.New("connection already exists")
	}
	if proc.State.State != 1 {
		return errors.New("process not is running")
	}
	if !proc.VerifyControl() {
		return errors.New("insufficient permissions")
	}
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return err
	}

	log.Logger.Infow("ws连接成功")
	data.UpdatedAt = time.Now()
	repository.WsShare.Edit(data)

	proc.SetTerminalSize(req.Cols, req.Rows)
	wsCtx, cancel := context.WithCancel(context.Background())
	wci := &WsConnetInstance{
		WsConnect:  conn,
		CancelFunc: cancel,
		wsLock:     sync.Mutex{},
	}
	proc.ReadCache(wci)
	w.startWsConnect(wci, cancel, proc, data.Write)
	proc.AddConn(guestName, wci)
	defer proc.DeleteConn(guestName)
	conn.SetCloseHandler(func(_ int, _ string) error {
		cancel()
		return nil
	})
	select {
	case <-proc.StopChan:
		log.Logger.Infow("ws连接断开", "操作类型", "进程已停止，强制断开ws连接")
	case <-time.After(time.Minute * time.Duration(config.CF.TerminalConnectTimeout)):
		log.Logger.Infow("ws连接断开", "操作类型", "连接时间超过最大时长限制")
	case <-wsCtx.Done():
		log.Logger.Infow("ws连接断开", "操作类型", "tcp连接建立已被关闭")
	case <-time.After(time.Until(data.ExpireTime)):
		log.Logger.Infow("ws连接断开", "操作类型", "分享时间已结束")
	}
	conn.Close()
	return
}

func (w *wsApi) startWsConnect(wci *WsConnetInstance, cancel context.CancelFunc, proc logic.Process, write bool) {
	log.Logger.Debugw("ws读取线程已启动")
	go func() {
		for {
			_, b, err := wci.WsConnect.ReadMessage()
			if err != nil {
				log.Logger.Debugw("ws读取线程已退出", "info", err)
				return
			}
			if write {
				proc.WriteBytes(b)
			}
		}
	}()

	// proactive health check
	pongChan := make(chan struct{})
	wci.WsConnect.SetPongHandler(func(appData string) error {
		pongChan <- struct{}{}
		return nil
	})
	timer := time.NewTicker(time.Second)
	go func() {
		defer timer.Stop()
		for {
			wci.wsLock.Lock()
			wci.WsConnect.WriteMessage(websocket.PingMessage, nil)
			wci.wsLock.Unlock()
			select {
			case <-timer.C:
				cancel()
				return
			case <-pongChan:
			}
			time.Sleep(time.Second * time.Duration(config.CF.WsHealthCheckInterval))
			timer.Reset(time.Second)
		}
	}()

}

func GetWsShareList(ctx *gin.Context, _ any) any {
	return logic.WsSahreLogic.GetWsShareList()
}

func DeleteWsShareById(ctx *gin.Context, _ any) any {
	return logic.WsSahreLogic.DeleteById(ctx.GetInt("id"))
}
