package app

import (
	"context"
	"sync"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/process"
	"github.com/lzh-1625/go_process_manager/log"
	"github.com/robfig/cron/v3"
	"go.uber.org/fx"
)

func NewApp(opts ...fx.Option) *fx.App {
	return fx.New(
		fx.NopLogger,
		fx.StopTimeout(time.Second*5+time.Duration(config.CF.KillWaitTime)*time.Second),
		Module,
		fx.Invoke(func(
			r *echo.Echo,
			lc fx.Lifecycle,
			processCtlLogic *logic.ProcessCtlLogic,
			taskLogic *logic.TaskLogic,
			logHandler *logic.LogHandler,
			eventLogic *logic.EventLogic,
		) {
			c := cron.New()
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go func() {
						log.Logger.Infow("starting echo server", "listen", config.CF.Listen)
						sc := echo.StartConfig{
							Address: config.CF.Listen,
						}
						err := sc.Start(ctx, r)
						if err != nil {
							log.Logger.Panicw("start echo server failed", "err", err)
						}
					}()
					// event cleaning cron job
					if config.CF.EventStorageTime >= 0 {
						c.AddFunc("0 3 * * *", func() {
							log.Logger.Infow("event cleaning execution")
							eventLogic.Clean(time.Duration(config.CF.EventStorageTime) * time.Hour * 24)
						})
						c.Start()
					}
					taskLogic.InitTaskJob()
					processCtlLogic.SetProcessStateHandler(taskLogic.RunTaskByTriggerEvent)
					processCtlLogic.ProcessInit()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					c.Stop()
					log.Logger.Infow("waiting for all process to stop")
					wg := sync.WaitGroup{}
					processCtlLogic.ForEach(func(proc *process.Process) {
						wg.Go(func() { proc.Kill() })
					})
					wg.Wait()
					logHandler.Close()
					return nil
				},
			})
		}),
		fx.Options(opts...),
	)
}
