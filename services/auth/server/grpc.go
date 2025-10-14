package server

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
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
	return &grpcServer{addr: addr}
}

func (s *grpcServer) Run() error {

	// Try common .env locations; don't exit fatally if not found.
	// Order: current dir, parent dir, filesystem root.
	// TECH DEBT
	if err := godotenv.Load(".env"); err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			if err := godotenv.Load("/.env"); err != nil {
				log.Println("Warning: .env not found at .env, ../.env, or /.env; relying on existing environment")
			}
		}
	}

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
		UserSvcAddr: os.Getenv("USER_SVC_PORT"),
	}

	authService := service.NewAuthService(
		googleConf,
		addConf,
	)

	handlers.NewGrpcAuthService(
		grpcServer,
		authService,
	)

	log.Println("Starting gRPC Server on", s.addr)

	return grpcServer.Serve(lis)
}
