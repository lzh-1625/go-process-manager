package middle

import (
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
	reqUser := c.GetHeader("Uuid")
	defer p.wc.TimeMap.Store(reqUser, p.wc.Ts)
	if ts, ok := p.wc.TimeMap.Load(reqUser); !ok || ts.(int64) > p.wc.Ts {
		c.Next()
		return
	}
	p.wc.Cond.L.Lock()
	defer p.wc.Cond.L.Unlock()
	p.wc.Cond.Wait()
	c.Next()
}

func (p *WaitCondMiddle) WaitTriggerMiddel(c *gin.Context) {
	defer p.Trigger()
	c.Next()
}
