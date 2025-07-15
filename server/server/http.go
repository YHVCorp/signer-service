package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/YHVCorp/signer-service/server/config"
	"github.com/YHVCorp/signer-service/server/utils"
	"github.com/gin-gonic/gin"
)

type FileManager struct {
	files       map[string]*FileInfo
	mu          sync.RWMutex
	uploadDir   string
	downloadDir string
}

type FileInfo struct {
	ID          string `json:"id"`
	OriginalURL string `json:"original_url,omitempty"`
	SignedURL   string `json:"signed_url,omitempty"`
	Status      string `json:"status"` // "uploaded", "signing", "signed", "ready"
}

type UploadResponse struct {
	FileID string `json:"file_id"`
}

type StatusResponse struct {
	Status    string `json:"status"`
	SignedURL string `json:"signed_url,omitempty"`
}

func NewFileManager() *FileManager {
	basePath := utils.GetMyPath()
	return &FileManager{
		files:       make(map[string]*FileInfo),
		uploadDir:   filepath.Join(basePath, "uploads"),
		downloadDir: filepath.Join(basePath, "downloads"),
	}
}

func (fm *FileManager) Setup() error {
	if err := os.MkdirAll(fm.uploadDir, 0755); err != nil {
		return err
	}
	return os.MkdirAll(fm.downloadDir, 0755)
}

func (fm *FileManager) generateFileID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (fm *FileManager) SetupHTTPRoutes(router *gin.Engine, signerServer *SignerServer) {
	api := router.Group("/api/v1")
	api.Use(fm.authMiddleware())

	api.POST("/upload", fm.uploadFile(signerServer))
	api.GET("/status/:file_id", fm.getFileStatus)
	api.GET("/download/:file_id", fm.downloadSignedFile)
	api.POST("/upload-signed/:file_id", fm.uploadSignedFile)

	router.GET("/unsigned/:file_id", fm.downloadUnsignedFile)
}

func (fm *FileManager) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization token"})
			c.Abort()
			return
		}

		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		expectedToken, err := config.GetDecryptedToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "server configuration error"})
			c.Abort()
			return
		}

		if token != expectedToken {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (fm *FileManager) uploadFile(signerServer *SignerServer) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, _, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to get file"})
			return
		}
		defer file.Close()

		fileID := fm.generateFileID()
		filePath := filepath.Join(fm.uploadDir, fileID)

		outFile, err := os.Create(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
			return
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
			return
		}

		fileInfo := &FileInfo{
			ID:          fileID,
			OriginalURL: fmt.Sprintf("/unsigned/%s", fileID),
			Status:      "uploaded",
		}

		fm.mu.Lock()
		fm.files[fileID] = fileInfo
		fm.mu.Unlock()

		downloadURL := fmt.Sprintf("http://%s/unsigned/%s", c.Request.Host, fileID)
		uploadURL := fmt.Sprintf("http://%s/api/v1/upload-signed/%s", c.Request.Host, fileID)

		fileInfo.Status = "signing"
		signerServer.SendSignRequest("default", fileID, downloadURL, uploadURL)

		c.JSON(http.StatusOK, UploadResponse{FileID: fileID})
	}
}

func (fm *FileManager) getFileStatus(c *gin.Context) {
	fileID := c.Param("file_id")

	fm.mu.RLock()
	fileInfo, exists := fm.files[fileID]
	fm.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	response := StatusResponse{
		Status: fileInfo.Status,
	}

	if fileInfo.Status == "ready" && fileInfo.SignedURL != "" {
		response.SignedURL = fmt.Sprintf("http://%s/api/v1/download/%s", c.Request.Host, fileID)
	}

	c.JSON(http.StatusOK, response)
}

func (fm *FileManager) downloadUnsignedFile(c *gin.Context) {
	fileID := c.Param("file_id")
	filePath := filepath.Join(fm.uploadDir, fileID)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.File(filePath)
}

func (fm *FileManager) uploadSignedFile(c *gin.Context) {
	fileID := c.Param("file_id")

	fm.mu.Lock()
	fileInfo, exists := fm.files[fileID]
	if !exists {
		fm.mu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	fm.mu.Unlock()

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to get file"})
		return
	}
	defer file.Close()

	signedFilePath := filepath.Join(fm.downloadDir, fileID)
	outFile, err := os.Create(signedFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save signed file"})
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save signed file"})
		return
	}

	fm.mu.Lock()
	fileInfo.Status = "ready"
	fileInfo.SignedURL = signedFilePath
	fm.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{"status": "signed file uploaded successfully"})
}

func (fm *FileManager) downloadSignedFile(c *gin.Context) {
	fileID := c.Param("file_id")

	fm.mu.RLock()
	fileInfo, exists := fm.files[fileID]
	fm.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	if fileInfo.Status != "ready" {
		c.JSON(http.StatusNotFound, gin.H{"error": "signed file not ready"})
		return
	}

	signedFilePath := filepath.Join(fm.downloadDir, fileID)
	if _, err := os.Stat(signedFilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "signed file not found"})
		return
	}

	c.File(signedFilePath)

	go fm.cleanupFile(fileID)
}

func (fm *FileManager) cleanupFile(fileID string) {
	fm.mu.Lock()
	delete(fm.files, fileID)
	fm.mu.Unlock()

	os.Remove(filepath.Join(fm.uploadDir, fileID))
	os.Remove(filepath.Join(fm.downloadDir, fileID))
}
