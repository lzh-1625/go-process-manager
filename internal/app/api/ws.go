package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/process"
	"github.com/lzh-1625/go_process_manager/log"

	"github.com/gorilla/websocket"
)

type WsApi struct {
	processCtlLogic *logic.ProcessCtlLogic
	wsShareLogic    *logic.WsShareLogic
	eventLogic      *logic.EventLogic
	permissionApi   *PermissionApi
}

func NewWsApi(
	processCtlLogic *logic.ProcessCtlLogic,
	wsShareLogic *logic.WsShareLogic,
	eventLogic *logic.EventLogic,
	permissionApi *PermissionApi) *WsApi {
	return &WsApi{
		processCtlLogic: processCtlLogic,
		wsShareLogic:    wsShareLogic,
		eventLogic:      eventLogic,
		permissionApi:   permissionApi,
	}
}

type WsConnetInstance struct {
	WsConnect  *websocket.Conn
	wsLock     sync.Mutex
	CancelFunc context.CancelFunc
}

func (w *WsConnetInstance) Write(b []byte) {
	w.wsLock.Lock()
	defer w.wsLock.Unlock()
	w.WsConnect.WriteMessage(websocket.TextMessage, b)
}

func (w *WsConnetInstance) Cancel() {
	w.CancelFunc()
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有跨域请求
	},
}

func (w *WsApi) WebsocketHandle(ctx *echo.Context) (err error) {
	var req model.WebsocketHandleReq
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	if !w.permissionApi.HasOprPermission(ctx, req.UUID, eum.OperationTerminal) {
		return errors.New("not permission")
	}
	reqUser := getUserName(ctx)
	proc, err := w.processCtlLogic.GetProcess(req.UUID)
	if err != nil {
		return err
	}
	if proc.HasWsConn(reqUser) {
		return errors.New("connection already exists")
	}
	if proc.Control.Controller != reqUser && !proc.VerifyControl() {
		return errors.New("insufficient permissions")
	}
	conn, err := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()
	log.Logger.Infow("ws connection success", "user", reqUser, "process", proc.Name)

	wsCtx, cancel := context.WithCancel(context.Background())
	wci := &WsConnetInstance{
		WsConnect:  conn,
		CancelFunc: cancel,
		wsLock:     sync.Mutex{},
	}
	if err := proc.ReadCache(wci); err != nil {
		return nil
	}
	if proc.State.State == eum.ProcessStateRunning {
		proc.SetTerminalSize(req.Cols, req.Rows)
		write := w.permissionApi.HasOprPermission(ctx, req.UUID, eum.OperationTerminalWrite)
		w.eventLogic.Create(proc.Name, eum.EventProcessConnect, "user", reqUser, "write", strconv.FormatBool(write))
		w.startWsConnect(wci, cancel, proc, write)
		proc.AddConn(reqUser, wci)
		defer proc.DeleteConn(reqUser)
	}
	conn.SetCloseHandler(func(_ int, _ string) error {
		cancel()
		return nil
	})
	select {
	case <-time.After(time.Minute * time.Duration(config.CF.TerminalConnectTimeout)):
		log.Logger.Infow("ws connection closed", "reason", "connection timeout")
	case <-wsCtx.Done():
		log.Logger.Infow("ws connection closed", "reason", "tcp connection closed")
	}
	return
}

func (w *WsApi) WebsocketShareHandle(ctx *echo.Context) (err error) {
	var req model.WebsocketHandleReq
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	data, err := w.wsShareLogic.GetWsShareDataByToken(req.Token)
	if err != nil {
		return err
	}
	if data.ExpireTime.Before(time.Now()) {
		return errors.New("share expired")
	}
	proc, err := w.processCtlLogic.GetProcess(data.Pid)
	if err != nil {
		return err
	}
	guestName := "guest-" + strconv.Itoa(int(data.ID)) // construct guest username
	if proc.HasWsConn(guestName) {
		return errors.New("connection already exists")
	}
	if proc.State.State != eum.ProcessStateRunning {
		return errors.New("process not is running")
	}
	if !proc.VerifyControl() {
		return errors.New("insufficient permissions")
	}
	conn, err := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()
	log.Logger.Infow("ws connection success", "user", data.CreateBy, "process", proc.Name)
	data.LastLink = time.Now()
	w.wsShareLogic.Edit(data)

	proc.SetTerminalSize(req.Cols, req.Rows)
	wsCtx, cancel := context.WithCancel(ctx.Request().Context())
	wci := &WsConnetInstance{
		WsConnect:  conn,
		CancelFunc: cancel,
		wsLock:     sync.Mutex{},
	}
	if err := proc.ReadCache(wci); err != nil {
		return nil
	}
	w.eventLogic.Create(proc.Name, eum.EventProcessConnect, "user", guestName, "by", data.CreateBy, "write", strconv.FormatBool(data.Write))
	w.startWsConnect(wci, cancel, proc, data.Write)
	proc.AddConn(guestName, wci)
	defer proc.DeleteConn(guestName)
	conn.SetCloseHandler(func(_ int, _ string) error {
		cancel()
		return nil
	})
	select {
	case <-proc.StopChan:
		log.Logger.Infow("ws connection closed", "reason", "process stopped, force close ws connection")
	case <-time.After(time.Minute * time.Duration(config.CF.TerminalConnectTimeout)):
		log.Logger.Infow("ws connection closed", "reason", "connection timeout")
	case <-wsCtx.Done():
		log.Logger.Infow("ws connection closed", "reason", "tcp connection closed")
	case <-time.After(time.Until(data.ExpireTime)):
		log.Logger.Infow("ws connection closed", "reason", "share time expired")
	}
	return
}

func (w *WsApi) startWsConnect(wci *WsConnetInstance, cancel context.CancelFunc, proc *process.ProcessPty, write bool) {
	log.Logger.Debugw("ws read thread started")
	go func() {
		for {
			_, b, err := wci.WsConnect.ReadMessage()
			if err != nil {
				log.Logger.Debugw("ws read thread exited", "info", err)
				cancel()
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

func (w *WsApi) GetWsShareList(ctx *echo.Context) error {
	return ctx.JSON(http.StatusOK, model.Response[[]*model.WsShare]{
		Data:    w.wsShareLogic.GetWsShareList(),
		Message: "success",
		Code:    0,
	})
}

func (w *WsApi) DeleteWsShareByID(ctx *echo.Context) error {
	var req struct {
		ID int `query:"id"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return w.wsShareLogic.DeleteByID(req.ID)
}
