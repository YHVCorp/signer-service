package utils

import (
	"fmt"
	"os/exec"
	"runtime"
)

func Execute(c string, dir string, arg ...string) error {
	cmd := exec.Command(c, arg...)
	cmd.Dir = dir

	return cmd.Run()
}

func ExecuteSignTool(certPath, key, container, filePath string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("signtool is only supported on Windows")
	}

	// Construct the command to execute signtool
	// Example command:
	// signtool sign /fd SHA256 /tr http://timestamp.digicert.com /td SHA256 /f "<CERT>" /csp "eToken Base Cryptographic Provider" /k "[{{<KEY>}}]=<CONTAINER>" "<FILE_TO_SIGN>"
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

	err := Execute("signtool", GetMyPath(), args...)
	if err != nil {
		return fmt.Errorf("signtool failed: %v", err)
	}

	return nil
}
