package utils

import (
	"fmt"
	"os/exec"
	"runtime"
)

func Execute(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func ExecuteSignTool(certPath, key, container, filePath string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("signtool is only supported on Windows")
	}

	args := []string{
		"sign",
		"/fd", "SHA256",
		"/tr", "http://timestamp.digicert.com",
		"/td", "SHA256",
		"/f", certPath,
		"/csp", "eToken Base Cryptographic Provider",
		"/k", fmt.Sprintf("[{{%s}}]=%s", key, container),
		filePath,
	}

	output, err := Execute("signtool", args...)
	if err != nil {
		return fmt.Errorf("signtool failed: %v\nOutput: %s", err, output)
	}

	return nil
}
