package serv

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/YHVCorp/signer-service/client/config"
	"github.com/YHVCorp/signer-service/client/utils"
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

	if !config.ConfigExists() {
		utils.Logger.Fatal("Configuration not found. Please create one using the -install flag.")
	}

	cfg, err := config.GetDecryptedConfig()
	if err != nil {
		utils.Logger.Fatal("Failed to load configuration: %v", err)
	}

	client := NewSignerClient(cfg.ServerAddress, cfg.Token, cfg.CertPath, cfg.Key, cfg.Container)
	err = client.Start()
	if err != nil {
		utils.Logger.Fatal("Failed to start signer client: %v", err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
}
