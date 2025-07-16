package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/YHVCorp/signer-service/client/config"
	"github.com/YHVCorp/signer-service/client/serv"
	"github.com/YHVCorp/signer-service/client/utils"
)

func main() {
	utils.InitLogger(config.ServiceLogFile)

	var (
		install      = flag.Bool("install", false, "Install the signer client service")
		uninstall    = flag.Bool("uninstall", false, "Uninstall the signer client service")
		run          = flag.Bool("run", false, "Run the signer client service")
		help         = flag.Bool("help", false, "Show help")
		setToken     = flag.String("set-token", "", "Set authentication token")
		setCert      = flag.String("set-cert", "", "Set signing certificate path")
		setKey       = flag.String("set-key", "", "Set signing key")
		setContainer = flag.String("set-container", "", "Set signing container")
		setServer    = flag.String("set-server", "", "Set server address")
	)
	flag.Parse()

	switch {
	case *install:
		if !config.ConfigExists() {
			fmt.Println("Configuration not found. Let's create one.")
			if err := config.CreateInitialConfig(); err != nil {
				log.Fatalf("Failed to create configuration: %v", err)
			}
		}
		serv.InstallService()
		fmt.Println("Service installed successfully")

	case *uninstall:
		serv.UninstallService()
		fmt.Println("Service uninstalled successfully")

	case *run:
		serv.RunService()

	case *help:
		fmt.Println("Use the following flags to configure the service:")
		fmt.Println("  -set-token <token>        Set authentication token")
		fmt.Println("  -set-cert <path>          Set signing certificate path")
		fmt.Println("  -set-key <key>            Set signing key")
		fmt.Println("  -set-container <container> Set signing container")
		fmt.Println("  -set-server <address>     Set server address")

	case *setToken != "":
		if err := config.UpdateToken(*setToken); err != nil {
			log.Fatalf("Failed to set token: %v", err)
		}
		fmt.Println("Token updated successfully")

	case *setCert != "":
		if err := config.UpdateCertPath(*setCert); err != nil {
			log.Fatalf("Failed to set certificate path: %v", err)
		}
		fmt.Println("Certificate path updated successfully")

	case *setKey != "":
		if err := config.UpdateKey(*setKey); err != nil {
			log.Fatalf("Failed to set key: %v", err)
		}
		fmt.Println("Key updated successfully")

	case *setContainer != "":
		if err := config.UpdateContainer(*setContainer); err != nil {
			log.Fatalf("Failed to set container: %v", err)
		}
		fmt.Println("Container updated successfully")

	case *setServer != "":
		if err := config.UpdateServerAddress(*setServer); err != nil {
			log.Fatalf("Failed to set server address: %v", err)
		}
		fmt.Println("Server address updated successfully")

	default:
		flag.Usage()
		os.Exit(1)
	}
}
