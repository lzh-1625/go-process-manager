//go:build linux

package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/lzh-1625/go_process_manager/config"
	"github.com/spf13/cobra"
)

var dockerCmd = &cobra.Command{
	Use:   "docker CONTAINER",
	Short: "Run and supervise a Docker container",
	Long: `Start a Docker container if it is not already running, stream its logs,
and stop it gracefully when gpm receives an interrupt or termination signal.`,
	Example: "  gpm docker my-container",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		superviseContainer(cmd.Context(), args[0])
	},
}

func init() {
	rootCmd.AddCommand(dockerCmd)
}

func superviseContainer(ctx context.Context, containerName string) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	info, err := cli.ContainerInspect(ctx, containerName)
	if err != nil {
		panic(err)
	}
	var opt = container.LogsOptions{
		ShowStdout: true,
		Follow:     true,
	}
	if info.State.Status != container.StateRunning {
		opt.Since = strconv.FormatInt(time.Now().Unix(), 10)
		_ = cli.ContainerStart(ctx, containerName, container.StartOptions{})
	} else {
		opt.Tail = "50"
	}
	out, err := cli.ContainerLogs(ctx, containerName, opt)
	if err != nil {
		panic(err)
	}

	go func() {
		stdcopy.StdCopy(os.Stdout, os.Stderr, out)
		os.Exit(0)
	}()
	ch := make(chan os.Signal, 1)
	signal.Notify(
		ch,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	<-ch
	cli.ContainerStop(ctx, containerName, container.StopOptions{
		Timeout: new(config.CF.KillWaitTime),
	})
}
