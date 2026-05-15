package middle

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
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

func (p *WaitCondMiddle) WaitGetMiddel(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		version, err := strconv.ParseInt(c.Request().Header.Get("Version"), 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, model.Response[struct{}]{
				Code:    -1,
				Message: "version is invalid",
			})
		}
		ctx, cancel := context.WithTimeout(c.Request().Context(), time.Second*time.Duration(config.CF.CondWaitTime))
		defer cancel()

		p.wc.Wait(ctx, version)

		c.Response().Header().Set("Version", strconv.FormatInt(p.wc.Version.Load(), 10))
		return next(c)
	}
}

func (p *WaitCondMiddle) WaitTriggerMiddel(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		defer p.Trigger()
		return next(c)
	}
}
