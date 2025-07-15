package serv

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	pb "github.com/YHVCorp/signer-service/proto"
	"github.com/YHVCorp/signer-service/client/utils"
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
}

func NewSignerClient(serverAddress, token, certPath, key, container string) *SignerClient {
	return &SignerClient{
		serverAddress: serverAddress,
		token:         token,
		certPath:      certPath,
		key:           key,
		container:     container,
	}
}

func (c *SignerClient) Start() error {
	// Establish gRPC connection
	conn, err := grpc.Dial(c.serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	c.conn = conn
	c.client = pb.NewSignerServiceClient(conn)

	log.Printf("Connected to signer server at %s", c.serverAddress)

	// Start listening for sign requests
	return c.listenForSignRequests()
}

func (c *SignerClient) Stop() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *SignerClient) listenForSignRequests() error {
	ctx := context.Background()
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+c.token)

	stream, err := c.client.StreamSignRequests(ctx, &pb.Empty{})
	if err != nil {
		return fmt.Errorf("failed to start stream: %v", err)
	}

	log.Println("Listening for sign requests...")

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			log.Println("Stream ended")
			break
		}
		if err != nil {
			log.Printf("Error receiving request: %v", err)
			time.Sleep(5 * time.Second) // Wait before reconnecting
			continue
		}

		log.Printf("Received sign request for file: %s", req.FileName)
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

	// Download file
	filePath := filepath.Join(tempDir, req.FileName)
	if err := c.downloadFile(req.DownloadUrl, filePath); err != nil {
		log.Printf("Failed to download file: %v", err)
		c.reportError(req.RequestId, fmt.Sprintf("Download failed: %v", err))
		return
	}

	log.Printf("Downloaded file: %s", filePath)

	// Sign file
	if err := c.signFile(filePath); err != nil {
		log.Printf("Failed to sign file: %v", err)
		c.reportError(req.RequestId, fmt.Sprintf("Signing failed: %v", err))
		return
	}

	log.Printf("Successfully signed file: %s", filePath)

	// Upload signed file
	if err := c.uploadFile(req.UploadUrl, filePath); err != nil {
		log.Printf("Failed to upload file: %v", err)
		c.reportError(req.RequestId, fmt.Sprintf("Upload failed: %v", err))
		return
	}

	log.Printf("Successfully uploaded signed file: %s", filePath)

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

	req, err := http.NewRequest("PUT", url, file)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Authorization", "Bearer "+c.token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("upload failed with status: %s", resp.Status)
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
		log.Printf("Failed to report success: %v", err)
	}
}

func (c *SignerClient) reportError(requestID, errorMsg string) {
	ctx := context.Background()
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+c.token)

	_, err := c.client.ReportSignResult(ctx, &pb.SignResult{
		RequestId: requestID,
		Success:   false,
		Message:   errorMsg,
	})

	if err != nil {
		log.Printf("Failed to report error: %v", err)
	}
}
