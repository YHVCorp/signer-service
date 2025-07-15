package config

import (
	"path/filepath"

	"github.com/YHVCorp/signer-service/server/utils"
)

var (
	ServiceLogFile = filepath.Join(utils.GetMyPath(), "logs", "utmstack_agent.log")
)
