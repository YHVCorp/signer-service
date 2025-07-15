package serv

import "github.com/YHVCorp/signer-service/server/utils"

func UninstallService() {
	err := utils.StopService("SignerServiceServer")
	if err != nil {
		utils.Logger.Fatal("error stopping SignerServiceServer: %v", err)
	}
	err = utils.UninstallService("SignerServiceServer")
	if err != nil {
		utils.Logger.Fatal("error uninstalling SignerServiceServer: %v", err)
	}
}
