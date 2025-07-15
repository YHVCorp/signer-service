package serv

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/YHVCorp/signer-service/server/config"
	"github.com/YHVCorp/signer-service/server/server"
	"github.com/YHVCorp/signer-service/server/utils"
	"github.com/kardianos/service"
)

type program struct{}

func (p *program) Start(_ service.Service) error {
	go p.run()
	return nil
}

func (p *program) Stop(_ service.Service) error {
	return nil
}

func (p *program) run() {
	utils.InitLogger(config.ServiceLogFile)

	srv := server.NewServer()
	err := srv.Start("50052", "8081")
	if err != nil {
		utils.Logger.Fatal("error starting server: %v", err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
}
