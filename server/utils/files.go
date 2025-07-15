package utils

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func GetMyPath() string {
	ex, err := os.Executable()
	if err != nil {
		return ""
	}
	exPath := filepath.Dir(ex)
	return exPath
}

func ReadYAML(path string, result interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	d := yaml.NewDecoder(file)
	if err := d.Decode(result); err != nil {
		return err
	}

	return nil
}
