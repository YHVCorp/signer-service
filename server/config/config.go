package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	aesCrypt "github.com/AtlasInsideCorp/AtlasInsideAES"
	"github.com/YHVCorp/signer-service/server/utils"
	"gopkg.in/yaml.v2"
)

const (
	SaltSize       = 16
	ConfigFileName = "config.yaml"
)

type Config struct {
	Token string `yaml:"token"`
}

func GetConfigPath() string {
	return filepath.Join(utils.GetMyPath(), ConfigFileName)
}

func GenerateConfig() (string, error) {
	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("error generating token: %v", err)
	}
	token := base64.StdEncoding.EncodeToString(tokenBytes)

	// Generate salt
	salt := make([]byte, SaltSize)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("error generating salt: %v", err)
	}

	// Encrypt token with salt as key
	encrypted, err := aesCrypt.AESEncrypt(token, salt)
	if err != nil {
		return "", fmt.Errorf("error encrypting token: %v", err)
	}

	// Decode encrypted string to bytes and concatenate with salt
	encryptedBytes, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("error decoding encrypted data: %v", err)
	}
	final := append(encryptedBytes, salt...)
	encryptedToken := base64.StdEncoding.EncodeToString(final)

	// Save to YAML config file
	config := Config{Token: encryptedToken}
	configData, err := yaml.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("error marshaling config: %v", err)
	}

	configPath := GetConfigPath()
	err = os.WriteFile(configPath, configData, 0600)
	if err != nil {
		return "", fmt.Errorf("error writing config file: %v", err)
	}

	return token, nil
}

func GetConfig() (*Config, error) {
	configPath := GetConfigPath()
	var config Config

	if err := utils.ReadYAML(configPath, &config); err != nil {
		return nil, fmt.Errorf("error reading config: %v", err)
	}

	return &config, nil
}

func GetDecryptedToken() (string, error) {
	config, err := GetConfig()
	if err != nil {
		return "", err
	}

	return DecryptToken(config.Token)
}

func DecryptToken(encryptedToken string) (string, error) {
	// Decode base64
	data, err := base64.StdEncoding.DecodeString(encryptedToken)
	if err != nil {
		return "", fmt.Errorf("error decoding base64: %v", err)
	}

	if len(data) < SaltSize {
		return "", fmt.Errorf("data too short to contain salt")
	}

	// Separate encrypted data and salt
	encryptedBytes := data[:len(data)-SaltSize]
	salt := data[len(data)-SaltSize:]

	// Encode encrypted bytes back to string for AESDecrypt
	encryptedString := base64.StdEncoding.EncodeToString(encryptedBytes)

	// Decrypt using salt as key
	decrypted, err := aesCrypt.AESDecrypt(encryptedString, salt)
	if err != nil {
		return "", fmt.Errorf("error decrypting token: %v", err)
	}

	return decrypted, nil
}

func GenerateNewToken() (string, error) {
	// Remove existing config file
	configPath := GetConfigPath()
	if _, err := os.Stat(configPath); err == nil {
		os.Remove(configPath)
	}

	// Generate new config
	return GenerateConfig()
}
