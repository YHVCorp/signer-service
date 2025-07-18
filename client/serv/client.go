package serv

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/YHVCorp/signer-service/client/utils"
	pb "github.com/YHVCorp/signer-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type SignerClient struct {
	serverAddress string
	token         string
	certPath      string
	key           string
	container     string
	client        pb.SignerServiceClient
	conn          *grpc.ClientConn

	maxRetries int
	retryDelay time.Duration
	isRunning  bool
}

func NewSignerClient(serverAddress, token, certPath, key, container string) *SignerClient {
	return &SignerClient{
		serverAddress: serverAddress,
		token:         token,
		certPath:      certPath,
		key:           key,
		container:     container,
		maxRetries:    -1,
		retryDelay:    1 * time.Second,
		isRunning:     false,
	}
}

func (c *SignerClient) Start() error {
	c.isRunning = true
	return c.connectWithRetry()
}

func (c *SignerClient) Stop() {
	c.isRunning = false
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *SignerClient) connectWithRetry() error {
	retryCount := 0

	for c.isRunning {
		err := c.connect()
		if err == nil {
			err = c.listenForSignRequests()
			if err != nil {
				log.Printf("Listen error: %v", err)
			}
		}

		if !c.isRunning {
			break
		}

		retryCount++
		log.Printf("Connection lost (attempt %d). Retrying in %v...", retryCount, c.retryDelay)

		if c.maxRetries > 0 && retryCount >= c.maxRetries {
			return utils.Logger.ErrorF("max retries (%d) reached", c.maxRetries)
		}

		time.Sleep(c.retryDelay)

		if c.retryDelay < 30*time.Second {
			c.retryDelay = time.Duration(float64(c.retryDelay) * 1.5)
		}
	}

	return nil
}

func (c *SignerClient) connect() error {
	if c.conn != nil {
		c.conn.Close()
	}

	server := strings.TrimPrefix(c.serverAddress, "https://")
	server = strings.TrimPrefix(server, "http://")

	log.Printf("Connecting to gRPC server at %s:50052", server)

	conn, err := grpc.Dial(fmt.Sprintf("%s:50052", server), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return utils.Logger.ErrorF("failed to connect to server: %v", err)
	}

	c.conn = conn
	c.client = pb.NewSignerServiceClient(conn)

	utils.Logger.Info("Successfully connected to gRPC server at %s:50052", server)

	c.retryDelay = 1 * time.Second

	return nil
}

func (c *SignerClient) listenForSignRequests() error {
	ctx := context.Background()
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+c.token)

	stream, err := c.client.StreamSignRequests(ctx, &pb.Empty{})
	if err != nil {
		return utils.Logger.ErrorF("failed to start stream: %v", err)
	}

	utils.Logger.Info("Listening for sign requests...")

	for c.isRunning {
		req, err := stream.Recv()
		if err == io.EOF {
			log.Println("Stream ended by server")
			return utils.Logger.ErrorF("stream closed by server")
		}
		if err != nil {
			log.Printf("Error receiving request: %v", err)
			return utils.Logger.ErrorF("stream error: %v", err)
		}

		utils.Logger.Info("Received sign request for file: %s", req.FileName)
		go c.processSignRequest(req)
	}

	return nil
}

func (c *SignerClient) processSignRequest(req *pb.SignRequest) {
	// Create temporary directory for processing
	tempDir, err := os.MkdirTemp("", "signer-client-*")
	if err != nil {
		log.Printf("Failed to create temp directory: %v", err)
		return
	}
	defer os.RemoveAll(tempDir)

	downloadURL := fmt.Sprintf("%s:8081%s", c.serverAddress, req.DownloadUrl)
	uploadURL := fmt.Sprintf("%s:8081%s", c.serverAddress, req.UploadUrl)

	// Download file
	filePath := filepath.Join(tempDir, req.FileName)
	if err := c.downloadFile(downloadURL, filePath); err != nil {
		utils.Logger.ErrorF("Failed to download file: %v", err)
		c.reportError(req.RequestId, fmt.Sprintf("Download failed: %v", err))
		return
	}

	utils.Logger.Info("Downloaded file: %s", filePath)

	// Sign file
	if err := c.signFile(filePath); err != nil {
		utils.Logger.ErrorF("Failed to sign file: %v", err)
		c.reportError(req.RequestId, fmt.Sprintf("Signing failed: %v", err))
		return
	}

	utils.Logger.Info("Successfully signed file: %s", filePath)

	// Upload signed file
	if err := c.uploadFile(uploadURL, filePath); err != nil {
		utils.Logger.ErrorF("Failed to upload file: %v", err)
		c.reportError(req.RequestId, fmt.Sprintf("Upload failed: %v", err))
		return
	}

	utils.Logger.Info("Successfully uploaded signed file: %s", filePath)

	// Report success
	c.reportSuccess(req.RequestId)
}

func (c *SignerClient) downloadFile(url, filePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (c *SignerClient) uploadFile(url, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}

	writer.Close()

	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.token)

	client := &http.Client{
		Timeout: 5 * time.Minute,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return utils.Logger.ErrorF("upload failed with status: %s", resp.Status)
	}

	return nil
}

func (c *SignerClient) signFile(filePath string) error {
	return utils.ExecuteSignTool(c.certPath, c.key, c.container, filePath)
}

func (c *SignerClient) reportSuccess(requestID string) {
	ctx := context.Background()
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+c.token)

	_, err := c.client.ReportSignResult(ctx, &pb.SignResult{
		RequestId: requestID,
		Success:   true,
		Message:   "File signed successfully",
	})

	if err != nil {
		utils.Logger.ErrorF("Failed to report success: %v", err)
	}
}

func (c *SignerClient) reportError(requestID, errorMsg string) {
	if !c.isRunning {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+c.token)

	_, err := c.client.ReportSignResult(ctx, &pb.SignResult{
		RequestId: requestID,
		Success:   false,
		Message:   errorMsg,
	})

	if err != nil {
		utils.Logger.ErrorF("Failed to report error: %v", err)
	} else {
		utils.Logger.Info("Successfully reported error for request %s: %s", requestID, errorMsg)
	}
}
