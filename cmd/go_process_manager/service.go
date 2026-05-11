package main

import (
	"log"
	"os"

	"github.com/kardianos/service"
	"github.com/lzh-1625/go_process_manager/boot"
	"github.com/lzh-1625/go_process_manager/internal/app/route"
	"github.com/lzh-1625/go_process_manager/utils"
	"github.com/spf13/cobra"
)

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
	Short: "Service the service",
	Long:  `Service the service`,
}

var serviceInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the service",
	Long:  `Install the service`,
	Run:   serviceAction,
}

var serviceUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the service",
	Long:  `Uninstall the service`,
	Run:   serviceAction,
}

var serviceStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the service",
	Long:  `Start the service`,
	Run:   serviceAction,
}

var serviceStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the service",
	Long:  `Stop the service`,
	Run:   serviceAction,
}

var serviceRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the service",
	Long:  `Restart the service`,
	Run:   serviceAction,
}

func serviceAction(cmd *cobra.Command, args []string) {
	svc, err := service.New(&Service{}, &service.Config{
		Name:             "go_process_manager",
		DisplayName:      "Go Process Manager",
		Description:      "Go Process Manager service",
		WorkingDirectory: utils.UnwarpIgnore(os.Getwd()),
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

type Service struct{}

func (s *Service) Start(_ service.Service) error {
	go s.run()
	return nil
}

func (s *Service) run() {
	boot.Boot()
	print(startTitle)
	route.Route()
}

func (s *Service) Stop(_ service.Service) error {
	return nil
}
