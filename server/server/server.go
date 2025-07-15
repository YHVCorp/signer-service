package server

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

type Server struct {
	grpcServer   *grpc.Server
	httpServer   *http.Server
	signerServer *SignerServer
	fileManager  *FileManager
}

func NewServer() *Server {
	return &Server{
		signerServer: NewSignerServer(),
		fileManager:  NewFileManager(),
	}
}

func (s *Server) Start(grpcPort, httpPort string) error {
	if err := s.fileManager.Setup(); err != nil {
		return fmt.Errorf("failed to setup file manager: %v", err)
	}

	errChan := make(chan error, 2)

	go func() {
		errChan <- s.startGRPCServer(grpcPort)
	}()

	go func() {
		errChan <- s.startHTTPServer(httpPort)
	}()

	return <-errChan
}

func (s *Server) startGRPCServer(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %v", port, err)
	}

	s.grpcServer = grpc.NewServer()
	s.signerServer.RegisterGRPC(s.grpcServer)

	log.Printf("gRPC server starting on port %s", port)
	return s.grpcServer.Serve(lis)
}

func (s *Server) startHTTPServer(port string) error {
	router := gin.Default()
	s.fileManager.SetupHTTPRoutes(router, s.signerServer)

	s.httpServer = &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	log.Printf("HTTP server starting on port %s", port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
	if s.httpServer != nil {
		s.httpServer.Close()
	}
}
