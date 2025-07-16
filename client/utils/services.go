package utils

import (
	"fmt"
	"os"
	"runtime"
)

func CheckIfServiceIsInstalled(serv string) (bool, error) {
	path := GetMyPath()
	var err error
	switch runtime.GOOS {
	case "windows":
		err = Execute("sc", path, "query", serv)
	case "linux":
		err = Execute("systemctl", path, "status", serv)
	case "darwin":
		err = Execute("launchctl", path, "list", serv)
	default:
		return false, fmt.Errorf("operative system unknown")
	}

	return err == nil, nil
}

func StopService(name string) error {
	path := GetMyPath()
	switch runtime.GOOS {
	case "windows":
		err := Execute("sc", path, "stop", name)
		if err != nil {
			return fmt.Errorf("error stoping service: %v", err)
		}
	case "linux":
		err := Execute("systemctl", path, "stop", name)
		if err != nil {
			return fmt.Errorf("error stoping service: %v", err)
		}
	case "darwin":
		err := Execute("launchctl", path, "remove", name)
		if err != nil {
			return fmt.Errorf("error stopping macOS service: %v", err)
		}
	}
	return nil
}

func UninstallService(name string) error {
	path := GetMyPath()
	switch runtime.GOOS {
	case "windows":
		err := Execute("sc.exe", path, "delete", name)
		if err != nil {
			return fmt.Errorf("error uninstalling service: %v", err)
		}
	case "linux":
		err := Execute("systemctl", path, "disable", name)
		if err != nil {
			return fmt.Errorf("error uninstalling service: %v", err)
		}
		err = Execute("rm", "/etc/systemd/system/", "/etc/systemd/system/"+name+".service")
		if err != nil {
			return fmt.Errorf("error uninstalling service: %v", err)
		}
	case "darwin":
		err := Execute("launchctl", path, "remove", name)
		if err != nil {
			return fmt.Errorf("error uninstalling service: %v", err)
		}
		err = Execute("rm", "/Library/LaunchDaemons/"+name+".plist")
		if err != nil {
			return fmt.Errorf("error uninstalling service: %v", err)
		}
		err = Execute("rm", "/Users/"+os.Getenv("USER")+"/Library/LaunchAgents/"+name+".plist")
		if err != nil {
			return fmt.Errorf("error uninstalling service: %v", err)
		}

	}
	return nil
}
