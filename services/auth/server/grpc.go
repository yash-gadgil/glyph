package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/yash-gadgil/glyph/services/auth/db"
	"github.com/yash-gadgil/glyph/services/auth/handlers"
	"github.com/yash-gadgil/glyph/services/auth/service"
	"github.com/yash-gadgil/glyph/services/auth/types"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
)

type grpcServer struct {
	addr string
}

func NewGrpcServer(addr string) *grpcServer {
	return &grpcServer{
		addr: addr,
	}
}

func (s *grpcServer) Run() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	googleConf := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_KEY"),
		ClientSecret: os.Getenv("GOOGLE_SECRET"),
		RedirectURL:  fmt.Sprintf("http://localhost%s/auth/oauth/google/callback", os.Getenv("GATEWAY_SVC_PORT")),
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}

	addConf := &types.AddrConfig{
		UserSvcAddr: os.Getenv("USER_SERVICE_ADDR"),
	}

	ctx := context.Background()
	cache := db.InitCache(ctx)

	authService := service.NewAuthService(
		googleConf,
		addConf,
		cache,
	)
	defer authService.Close()

	handlers.NewGrpcAuthService(
		grpcServer,
		authService,
	)

	log.Println("Starting gRPC Server on", s.addr)

	return grpcServer.Serve(lis)
}
