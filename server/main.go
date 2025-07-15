package main

import (
	"fmt"
	"os"

	"github.com/YHVCorp/signer-service/server/config"
	"github.com/YHVCorp/signer-service/server/serv"
	"github.com/YHVCorp/signer-service/server/utils"
	"github.com/common-nighthawk/go-figure"
)

func main() {
	utils.InitLogger(config.ServiceLogFile)
	myFigure := figure.NewFigure("SignerService", "", true)
	myFigure.Print()

	if len(os.Args) > 1 {
		arg := os.Args[1]

		isInstalled, err := utils.CheckIfServiceIsInstalled("SignerServiceServer")
		if err != nil {
			fmt.Println("Error checking if service is installed: ", err)
			os.Exit(1)
		}
		if arg != "install" && !isInstalled {
			fmt.Println("SignerServiceServer service is not installed")
			os.Exit(1)
		} else if arg == "install" && isInstalled {
			fmt.Println("SignerServiceServer service is already installed")
			os.Exit(1)
		}

		switch arg {
		case "run":
			fmt.Println("Starting SignerService server ...")
			serv.RunService()
		case "install":
			fmt.Println("Installing SignerServiceServer service ...")
			fmt.Print("Configuring server ... ")
			token, err := config.GenerateConfig()
			if err != nil {
				fmt.Println("\nError generating config: ", err)
				os.Exit(1)
			}
			fmt.Println("[OK]")

			fmt.Print("Creating service ... ")
			serv.InstallService()
			fmt.Println("[OK]")

			fmt.Printf("SignerServiceServer service installed correctly. You can use the token: %s for authenticate clients\n", token)

		case "generate-new-token":
			fmt.Println("Generating new authentication token ...")
			token, err := config.GenerateNewToken()
			if err != nil {
				fmt.Printf("Error generating new token: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("New authentication token generated: %s\n", token)

		case "uninstall":
			fmt.Println("Uninstalling SignerServiceServer service ...")

			configPath := config.GetConfigPath()
			if _, err := os.Stat(configPath); err == nil {
				os.Remove(configPath)
			}

			serv.UninstallService()

			fmt.Println("[OK]")
			fmt.Println("SignerServiceServer service uninstalled correctly")
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
	fmt.Println("### SignerServiceServer CLI ###")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  signer_service_server <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  install                  Install the SignerServiceServer as a system service")
	fmt.Println("  run                      Run the SignerServiceServer in the foreground")
	fmt.Println("  generate-new-token       Generate a new authentication token")
	fmt.Println("  uninstall                Uninstall the SignerServiceServer service")
	fmt.Println("  help                     Display this help message")
	fmt.Println()
	fmt.Println("Notes:")
	fmt.Println("  - Run commands with appropriate privileges (e.g., administrator rights).")
	fmt.Println("  - Logs are stored in the service log file.")
	fmt.Println()
	os.Exit(0)
}
