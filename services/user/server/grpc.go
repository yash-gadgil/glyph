package server

import (
	"log"
	"net"

	"github.com/yash-gadgil/glyph/services/user/handler"
	"github.com/yash-gadgil/glyph/services/user/service"
	"google.golang.org/grpc"
)

type gRPCServer struct {
	addr string
}

func NewGRPCServer(addr string) *gRPCServer {
	return &gRPCServer{addr: addr}
}

func (s *gRPCServer) Run() error {

	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	watchlistService := service.NewWatchlistService()
	handler.NewGrpcWatchlistService(grpcServer, watchlistService)

	log.Println("Starting gRPC Server on", s.addr)

	return grpcServer.Serve(lis)
}
