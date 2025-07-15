package server

import (
	"context"
	"log"
	"sync"

	"github.com/YHVCorp/signer-service/proto"
	"google.golang.org/grpc"
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
			log.Printf("Sign request sent for file %s", fileName)
		default:
			log.Printf("Failed to send sign request for %s: channel full", fileName)
		}
	} else {
		log.Printf("Client %s not found", clientID)
	}
}

func (s *SignerServer) ReportSignResult(ctx context.Context, result *proto.SignResult) (*proto.Empty, error) {
	log.Printf("Received sign result for request %s: success=%t, message=%s",
		result.RequestId, result.Success, result.Message)
	return &proto.Empty{}, nil
}

func (s *SignerServer) RegisterGRPC(grpcServer *grpc.Server) {
	proto.RegisterSignerServiceServer(grpcServer, s)
}
