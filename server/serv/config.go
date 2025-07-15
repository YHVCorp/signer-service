package serv

import (
	"github.com/kardianos/service"
)

func GetConfigServ() *service.Config {
	svcConfig := &service.Config{
		Name:        "SignerServiceServer",
		DisplayName: "Signer Service Server",
		Description: "Signer Service Server",
		Arguments:   []string{"run"},
	}

	return svcConfig
}
