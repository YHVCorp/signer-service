package server

import (
	"sync"

	"github.com/YHVCorp/signer-service/server/proto"
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

func (s *SignerServer) Subscribe(req *proto.SubscribeRequest, stream proto.SignerService_SubscribeServer) error {
	clientID := req.ClientId
	if clientID == "" {
		clientID = "default"
	}

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

func (s *SignerServer) SendSignRequest(clientID, fileID, downloadURL, uploadURL string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if clientChan, exists := s.clients[clientID]; exists {
		signReq := &proto.SignRequest{
			FileId:      fileID,
			DownloadUrl: downloadURL,
			UploadUrl:   uploadURL,
		}

		select {
		case clientChan <- signReq:
		default:
		}
	}
}

func (s *SignerServer) RegisterGRPC(grpcServer *grpc.Server) {
	proto.RegisterSignerServiceServer(grpcServer, s)
}
