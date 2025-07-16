package serv

import "github.com/YHVCorp/signer-service/client/utils"

func UninstallService() {
	err := utils.StopService("SignerClientService")
	if err != nil {
		utils.Logger.ErrorF("error stopping SignerClientService: %v", err)
	}
	err = utils.UninstallService("SignerClientService")
	if err != nil {
		utils.Logger.ErrorF("error uninstalling SignerClientService: %v", err)
	}
}
