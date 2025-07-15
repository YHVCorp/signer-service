package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func ReadYAML(filename string, out interface{}) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, out)
}

func GetMyPath() string {
	ex, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(ex)
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func EnsureDir(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, 0755)
	}
	return nil
}

func TempFile(dir, pattern string) (*os.File, error) {
	return os.CreateTemp(dir, pattern)
}

func RemoveFile(filename string) error {
	if FileExists(filename) {
		return os.Remove(filename)
	}
	return nil
}

func DownloadFile(url, filepath string) error {
	// This will be implemented with actual HTTP download logic
	return fmt.Errorf("download functionality not yet implemented")
}

func UploadFile(url, filepath string) error {
	// This will be implemented with actual HTTP upload logic
	return fmt.Errorf("upload functionality not yet implemented")
}