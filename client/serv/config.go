package serv

import (
	"github.com/kardianos/service"
)

func GetConfigServ() *service.Config {
	svcConfig := &service.Config{
		Name:        "SignerServiceClient",
		DisplayName: "Signer Service Client",
		Description: "Signer Service Client",
		Arguments:   []string{"run"},
	}

	return svcConfig
}
