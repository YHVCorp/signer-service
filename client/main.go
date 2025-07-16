package main

import (
	"fmt"
	"log"
	"os"

	"github.com/YHVCorp/signer-service/client/config"
	"github.com/YHVCorp/signer-service/client/serv"
	"github.com/YHVCorp/signer-service/client/utils"
)

func main() {
	utils.InitLogger(config.ServiceLogFile)

	if len(os.Args) > 1 {
		arg := os.Args[1]

		isInstalled, err := utils.CheckIfServiceIsInstalled("SignerServiceClient")
		if err != nil {
			fmt.Println("Error checking if service is installed: ", err)
			os.Exit(1)
		}
		if arg != "install" && !isInstalled {
			fmt.Println("SignerServiceClient service is not installed")
			os.Exit(1)
		} else if arg == "install" && isInstalled {
			fmt.Println("SignerServiceClient service is already installed")
			os.Exit(1)
		}

		switch arg {
		case "run":
			fmt.Println("Starting SignerServiceClient server ...")
			serv.RunService()
		case "install":
			if !config.ConfigExists() {
				fmt.Println("Configuration not found. Let's create one.")
				if err := config.CreateInitialConfig(); err != nil {
					log.Fatalf("Failed to create configuration: %v", err)
				}
			}
			serv.InstallService()
			fmt.Println("Service installed successfully")

		case "setToken":
			if err := config.UpdateToken(os.Args[2]); err != nil {
				log.Fatalf("Failed to set token: %v", err)
			}
			fmt.Println("Token updated successfully")

		case "setCert":
			if err := config.UpdateCertPath(os.Args[2]); err != nil {
				log.Fatalf("Failed to set certificate path: %v", err)
			}
			fmt.Println("Certificate path updated successfully")

		case "setKey":
			if err := config.UpdateKey(os.Args[2]); err != nil {
				log.Fatalf("Failed to set key: %v", err)
			}
			fmt.Println("Key updated successfully")

		case "setContainer":
			if err := config.UpdateContainer(os.Args[2]); err != nil {
				log.Fatalf("Failed to set container: %v", err)
			}
			fmt.Println("Container updated successfully")

		case "setServer":
			if err := config.UpdateServerAddress(os.Args[2]); err != nil {
				log.Fatalf("Failed to set server address: %v", err)
			}
			fmt.Println("Server address updated successfully")

		case "uninstall":
			serv.UninstallService()
			fmt.Println("Service uninstalled successfully")

		case "help":
			Help()
		default:
			fmt.Println("unknown option")
		}
	} else {
		Help()
	}
}

func Help() {
	fmt.Println("### SignerServiceClient CLI ###")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  signer_service_client <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  install                  Install the SignerServiceClient as a system service")
	fmt.Println("  run                      Run the SignerServiceClient in the foreground")
	fmt.Println("  setToken <token>         Set the authentication token for the service")
	fmt.Println("  setCert <path>           Set the signing certificate path")
	fmt.Println("  setKey <key>             Set the signing key for the service")
	fmt.Println("  setContainer <container> Set the signing container for the service")
	fmt.Println("  setServer <address>      Set the server address for the service")
	fmt.Println("  uninstall                Uninstall the SignerServiceClient service")
	fmt.Println("  help                     Display this help message")
	fmt.Println()
	fmt.Println("Notes:")
	fmt.Println("  - Run commands with appropriate privileges (e.g., administrator rights).")
	fmt.Println("  - Logs are stored in the service log file.")
	fmt.Println()
	os.Exit(0)
}
