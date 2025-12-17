package middle

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
)

type WaitCondMiddle struct {
	wc *logic.WaitCond
}

func NewWaitCond(wc *logic.WaitCond) *WaitCondMiddle {
	return &WaitCondMiddle{
		wc: wc,
	}
}

func (p *WaitCondMiddle) Trigger() {
	p.wc.Trigger()
}

func (p *WaitCondMiddle) WaitGetMiddel(c *gin.Context) {
	version, err := strconv.ParseInt(c.GetHeader("Version"), 10, 64)
	if err != nil {
		rErr(c, -1, "version is invalid", err)
		return
	}
	if version < p.wc.Version.Load() {
		c.Next()
		return
	}
	p.wc.Cond.L.Lock()
	defer p.wc.Cond.L.Unlock()
	p.wc.Cond.Wait()
	c.Header("Version", strconv.FormatInt(p.wc.Version.Load(), 10))
	c.Next()
}

func (p *WaitCondMiddle) WaitTriggerMiddel(c *gin.Context) {
	defer p.Trigger()
	c.Next()
}
