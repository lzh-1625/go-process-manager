package main

import (
	"log"
	"os"

	"github.com/kardianos/service"
	"github.com/lzh-1625/go_process_manager/boot"
	"github.com/lzh-1625/go_process_manager/internal/app/route"
	"github.com/lzh-1625/go_process_manager/utils"

	"github.com/gin-gonic/gin"
)

var startTitle = `
----------------------------------------------------------------------------
          _____                    _____                    _____          
         /\    \                  /\    \                  /\    \         
        /::\    \                /::\    \                /::\____\        
       /::::\    \              /::::\    \              /::::|   |        
      /::::::\    \            /::::::\    \            /:::::|   |        
     /:::/\:::\    \          /:::/\:::\    \          /::::::|   |        
    /:::/  \:::\    \        /:::/__\:::\    \        /:::/|::|   |        
   /:::/    \:::\    \      /::::\   \:::\    \      /:::/ |::|   |        
  /:::/    / \:::\    \    /::::::\   \:::\    \    /:::/  |::|___|______  
 /:::/    /   \:::\ ___\  /:::/\:::\   \:::\____\  /:::/   |::::::::\    \ 
/:::/____/  ___\:::|    |/:::/  \:::\   \:::|    |/:::/    |:::::::::\____\
\:::\    \ /\  /:::|____|\::/    \:::\  /:::|____|\::/    / ~~~~~/:::/    /
 \:::\    /::\ \::/    /  \/_____/\:::\/:::/    /  \/____/      /:::/    / 
  \:::\   \:::\ \/____/            \::::::/    /               /:::/    /  
   \:::\   \:::\____\               \::::/    /               /:::/    /   
    \:::\  /:::/    /                \::/____/               /:::/    /    
     \:::\/:::/    /                  ~~                    /:::/    /     
      \::::::/    /                                        /:::/    /      
       \::::/    /                                        /:::/    /       
        \::/____/                                         \::/    /        
                                                           \/____/         
----------------------------------------------------------------------------  
`

func main() {

	svc, err := service.New(&Service{}, &service.Config{
		Name:             "go_process_manager",
		DisplayName:      "Go Process Manager",
		Description:      "Go Process Manager service",
		WorkingDirectory: utils.UnwarpIgnore(os.Getwd()),
	})
	if err != nil {
		log.Panic(err)
	}

	var svcAction string
	if len(os.Args) > 1 {
		svcAction = os.Args[1]
	}
	switch svcAction {
	case "install", "uninstall", "start", "stop", "restart":
		err = service.Control(svc, svcAction)
		if err != nil {
			log.Panic(err)
		}
		log.Println(svcAction, "success")
	default:
		if err := svc.Run(); err != nil {
			log.Panic(err)
		}
	}
}

type Service struct{}

func (s *Service) Start(_ service.Service) error {
	go s.run()
	return nil
}

func (s *Service) run() {
	boot.Boot()
	print(startTitle)
	gin.SetMode(gin.ReleaseMode)
	route.Route()
}

func (s *Service) Stop(_ service.Service) error {
	return nil
}
