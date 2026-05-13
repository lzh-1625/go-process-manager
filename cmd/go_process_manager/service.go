package main

import (
	"log"
	"os"

	"github.com/kardianos/service"
	"github.com/lzh-1625/go_process_manager/utils"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serviceCmd)
}

func init() {
	serviceCmd.AddCommand(
		serviceInstallCmd,
		serviceUninstallCmd,
		serviceStartCmd,
		serviceStopCmd,
		serviceRestartCmd,
	)
}

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage gpm as a system service",
	Long: `Manage the gpm system service (systemd on Linux, launchd on macOS,
Service Control Manager on Windows).

Use these sub-commands to install gpm so it starts automatically at boot,
or to start/stop/restart the background service from the command line.`,
	Example: `  gpm service install      # register gpm as a system service
  gpm service start        # start the service
  gpm service restart      # restart the service`,
}

var serviceInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install gpm as a system service",
	Long: `Register gpm as a system service using the current working directory
as the service's working directory. After installing, use "gpm service start"
to launch it.`,
	Run: serviceAction,
}

var serviceUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the gpm system service",
	Long:  `Remove the gpm service registration from the operating system.`,
	Run:   serviceAction,
}

var serviceStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the gpm system service",
	Long:  `Start the previously installed gpm system service in the background.`,
	Run:   serviceAction,
}

var serviceStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the gpm system service",
	Long:  `Gracefully stop the running gpm system service.`,
	Run:   serviceAction,
}

var serviceRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the gpm system service",
	Long:  `Stop and start the gpm system service. Useful after upgrading or changing the configuration.`,
	Run:   serviceAction,
}

func serviceAction(cmd *cobra.Command, args []string) {
	svc, err := service.New(&Service{cmd, args}, &service.Config{
		Name:             "gpm",
		DisplayName:      "Go Process Manager",
		Description:      "Go Process Manager service",
		WorkingDirectory: utils.UnwarpIgnore(os.UserHomeDir()),
		Arguments:        []string{"run"},
	})
	if err != nil {
		log.Panic(err)
	}

	err = service.Control(svc, cmd.Use)
	if err != nil {
		log.Panic(err)
	}
	log.Println(cmd.Use, "success")

}

type Service struct {
	cmd  *cobra.Command
	args []string
}

func (s *Service) Start(_ service.Service) error {
	go s.run()
	return nil
}

func (s *Service) run() {
	runCmd.Run(s.cmd, s.args)
}

func (s *Service) Stop(_ service.Service) error {
	return nil
}
