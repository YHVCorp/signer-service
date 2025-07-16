package server

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/YHVCorp/signer-service/proto"
	"github.com/YHVCorp/signer-service/server/config"
	"github.com/YHVCorp/signer-service/server/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type SignerServer struct {
	proto.UnimplementedSignerServiceServer
	clients map[string]chan *proto.SignRequest
	mu      sync.RWMutex
}

func NewSignerServer() *SignerServer {
	return &SignerServer{
		clients: make(map[string]chan *proto.SignRequest),
	}
}

func (s *SignerServer) StreamSignRequests(req *proto.Empty, stream proto.SignerService_StreamSignRequestsServer) error {
	if err := s.validateToken(stream.Context()); err != nil {
		utils.Logger.ErrorF("authentication failed: %v", err)
		return fmt.Errorf("authentication failed: %v", err)
	}

	clientID := "default"

	clientChan := make(chan *proto.SignRequest, 100)

	s.mu.Lock()
	s.clients[clientID] = clientChan
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, clientID)
		close(clientChan)
		s.mu.Unlock()
	}()

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case signReq := <-clientChan:
			if err := stream.Send(signReq); err != nil {
				return err
			}
		}
	}
}

func (s *SignerServer) SendSignRequest(clientID, requestID, fileName, downloadURL, uploadURL string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if clientChan, exists := s.clients[clientID]; exists {
		signReq := &proto.SignRequest{
			RequestId:   requestID,
			FileName:    fileName,
			DownloadUrl: downloadURL,
			UploadUrl:   uploadURL,
		}

		select {
		case clientChan <- signReq:
			utils.Logger.Info("Sign request sent for file %s", fileName)
		default:
			utils.Logger.ErrorF("Failed to send sign request for %s: channel full", fileName)
		}
	} else {
		utils.Logger.ErrorF("Client %s not found", clientID)
	}
}

func (s *SignerServer) validateToken(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return fmt.Errorf("no metadata found")
	}

	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return fmt.Errorf("no authorization header")
	}

	token := authHeader[0]
	token = strings.TrimPrefix(token, "Bearer ")

	expectedToken, err := config.GetDecryptedToken()
	if err != nil {
		return fmt.Errorf("server configuration error: %v", err)
	}

	if token != expectedToken {
		return fmt.Errorf("invalid token")
	}

	return nil
}

func (s *SignerServer) ReportSignResult(ctx context.Context, result *proto.SignResult) (*proto.Empty, error) {
	if err := s.validateToken(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %v", err)
	}

	utils.Logger.Info("Received sign result for request %s: success=%t, message=%s",
		result.RequestId, result.Success, result.Message)
	return &proto.Empty{}, nil
}

func (s *SignerServer) RegisterGRPC(grpcServer *grpc.Server) {
	proto.RegisterSignerServiceServer(grpcServer, s)
}
