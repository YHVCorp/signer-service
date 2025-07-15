package config

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	aesCrypt "github.com/AtlasInsideCorp/AtlasInsideAES"
	"github.com/YHVCorp/signer-service/client/utils"
	"golang.org/x/term"
	"gopkg.in/yaml.v2"
)

const (
	SaltSize       = 16
	ConfigFileName = "client-config.yaml"
)

type Config struct {
	Token         string `yaml:"token"`
	CertPath      string `yaml:"cert_path"`
	Key           string `yaml:"key"`
	Container     string `yaml:"container"`
	ServerAddress string `yaml:"server_address"`
}

func GetConfigPath() string {
	return filepath.Join(utils.GetMyPath(), ConfigFileName)
}

func encryptValue(value string) (string, error) {
	// Generate salt
	salt := make([]byte, SaltSize)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("error generating salt: %v", err)
	}

	// Encrypt value with salt as key
	encrypted, err := aesCrypt.AESEncrypt(value, salt)
	if err != nil {
		return "", fmt.Errorf("error encrypting value: %v", err)
	}

	// Decode encrypted string to bytes and concatenate with salt
	encryptedBytes, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("error decoding encrypted data: %v", err)
	}
	final := append(encryptedBytes, salt...)
	return base64.StdEncoding.EncodeToString(final), nil
}

func decryptValue(encryptedValue string) (string, error) {
	if encryptedValue == "" {
		return "", nil
	}

	// Decode base64
	data, err := base64.StdEncoding.DecodeString(encryptedValue)
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
		return "", fmt.Errorf("error decrypting value: %v", err)
	}

	return decrypted, nil
}

func GenerateConfig(token, certPath, key, container, serverAddress string) error {
	// Encrypt all sensitive values
	encryptedToken, err := encryptValue(token)
	if err != nil {
		return fmt.Errorf("error encrypting token: %v", err)
	}

	encryptedCertPath, err := encryptValue(certPath)
	if err != nil {
		return fmt.Errorf("error encrypting cert path: %v", err)
	}

	encryptedKey, err := encryptValue(key)
	if err != nil {
		return fmt.Errorf("error encrypting key: %v", err)
	}

	encryptedContainer, err := encryptValue(container)
	if err != nil {
		return fmt.Errorf("error encrypting container: %v", err)
	}

	encryptedServerAddress, err := encryptValue(serverAddress)
	if err != nil {
		return fmt.Errorf("error encrypting server address: %v", err)
	}

	// Save to YAML config file
	config := Config{
		Token:         encryptedToken,
		CertPath:      encryptedCertPath,
		Key:           encryptedKey,
		Container:     encryptedContainer,
		ServerAddress: encryptedServerAddress,
	}

	configData, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}

	configPath := GetConfigPath()
	err = os.WriteFile(configPath, configData, 0600)
	if err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}

	return nil
}

func GetConfig() (*Config, error) {
	configPath := GetConfigPath()
	var config Config

	if err := utils.ReadYAML(configPath, &config); err != nil {
		return nil, fmt.Errorf("error reading config: %v", err)
	}

	return &config, nil
}

func GetDecryptedConfig() (*DecryptedConfig, error) {
	config, err := GetConfig()
	if err != nil {
		return nil, err
	}

	token, err := decryptValue(config.Token)
	if err != nil {
		return nil, fmt.Errorf("error decrypting token: %v", err)
	}

	certPath, err := decryptValue(config.CertPath)
	if err != nil {
		return nil, fmt.Errorf("error decrypting cert path: %v", err)
	}

	key, err := decryptValue(config.Key)
	if err != nil {
		return nil, fmt.Errorf("error decrypting key: %v", err)
	}

	container, err := decryptValue(config.Container)
	if err != nil {
		return nil, fmt.Errorf("error decrypting container: %v", err)
	}

	serverAddress, err := decryptValue(config.ServerAddress)
	if err != nil {
		return nil, fmt.Errorf("error decrypting server address: %v", err)
	}

	return &DecryptedConfig{
		Token:         token,
		CertPath:      certPath,
		Key:           key,
		Container:     container,
		ServerAddress: serverAddress,
	}, nil
}

type DecryptedConfig struct {
	Token         string
	CertPath      string
	Key           string
	Container     string
	ServerAddress string
}

func UpdateToken(token string) error {
	return updateConfigField("token", token)
}

func UpdateCertPath(certPath string) error {
	return updateConfigField("cert_path", certPath)
}

func UpdateKey(key string) error {
	return updateConfigField("key", key)
}

func UpdateContainer(container string) error {
	return updateConfigField("container", container)
}

func UpdateServerAddress(serverAddress string) error {
	return updateConfigField("server_address", serverAddress)
}

func updateConfigField(field, value string) error {
	config, err := GetDecryptedConfig()
	if err != nil {
		return err
	}

	switch field {
	case "token":
		config.Token = value
	case "cert_path":
		config.CertPath = value
	case "key":
		config.Key = value
	case "container":
		config.Container = value
	case "server_address":
		config.ServerAddress = value
	default:
		return fmt.Errorf("unknown field: %s", field)
	}

	return GenerateConfig(config.Token, config.CertPath, config.Key, config.Container, config.ServerAddress)
}

func ConfigExists() bool {
	_, err := os.Stat(GetConfigPath())
	return err == nil
}

func CreateInitialConfig() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter server address (e.g., localhost:50051): ")
	serverAddress, _ := reader.ReadString('\n')
	serverAddress = strings.TrimSpace(serverAddress)

	fmt.Print("Enter authentication token: ")
	tokenBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read token: %v", err)
	}
	token := string(tokenBytes)
	fmt.Println()

	fmt.Print("Enter signing certificate path: ")
	certPath, _ := reader.ReadString('\n')
	certPath = strings.TrimSpace(certPath)

	// Validate certificate path
	if !filepath.IsAbs(certPath) {
		return fmt.Errorf("certificate path must be absolute")
	}
	if !utils.FileExists(certPath) {
		return fmt.Errorf("certificate file does not exist: %s", certPath)
	}

	fmt.Print("Enter signing key: ")
	keyBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read key: %v", err)
	}
	key := string(keyBytes)
	fmt.Println()

	fmt.Print("Enter signing container: ")
	containerBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read container: %v", err)
	}
	container := string(containerBytes)
	fmt.Println()

	return GenerateConfig(token, certPath, key, container, serverAddress)
}
