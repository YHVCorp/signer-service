package config

import (
	"path/filepath"

	"github.com/YHVCorp/signer-service/client/utils"
)

var (
	ServiceLogFile = filepath.Join(utils.GetMyPath(), "logs", "signer_agent.log")
)
