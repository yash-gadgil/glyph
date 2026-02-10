package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/pkg/telemetry"
	"github.com/yash-gadgil/glyph/services/auth/db"
	"github.com/yash-gadgil/glyph/services/auth/handlers"
	"github.com/yash-gadgil/glyph/services/auth/types"
	"github.com/yash-gadgil/glyph/services/auth/utils"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type grpcServer struct {
	addr string
	log  *zap.Logger
}

func NewGrpcServer(addr string) *grpcServer {
	return &grpcServer{addr: addr, log: logger.New("auth-service")}
}

func (s *grpcServer) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		s.log.Error("failed_to_listen", logger.KV("addr", s.addr), zap.Error(err))
		return err
	}

	telemetry.EnableGRPCHistograms()
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logger.UnaryServerInterceptor(s.log),
			grpc_prometheus.UnaryServerInterceptor,
		),
	)

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	oauthRedirectURL := os.Getenv("OAUTH_REDIRECT_URL")
	if oauthRedirectURL == "" {
		gatewayPort := os.Getenv("GATEWAY_SVC_PORT")
		if gatewayPort == "" {
			gatewayPort = ":8080"
		}
		oauthRedirectURL = fmt.Sprintf("http://localhost%s/auth/oauth/google/callback", gatewayPort)
	}

	googleConf := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_KEY"),
		ClientSecret: os.Getenv("GOOGLE_SECRET"),
		RedirectURL:  oauthRedirectURL,
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}

	addConf := &types.AddrConfig{
		UserSvcAddr: os.Getenv("USER_SERVICE_ADDR"),
	}

	cache := db.InitCache(ctx, s.log)

	keyStore, err := utils.NewKeyStore(s.log)
	if err != nil {
		s.log.Fatal("failed_to_initialize_keystore", zap.Error(err))
	}

	rotationStop := keyStore.StartRotationScheduler(30 * 24 * time.Hour)

	authHandler := handlers.NewAuthHandler(googleConf, addConf, cache, keyStore, s.log)
	defer authHandler.Close()

	handlers.Register(grpcServer, authHandler)

	grpc_prometheus.Register(grpcServer)
	go telemetry.ServeMetrics(ctx, telemetry.MetricsAddr(), s.log)

	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	s.log.Info("grpc_server_running", logger.KV("addr", s.addr))

	go func() {
		<-ctx.Done()
		healthServer.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
		s.log.Info("shutting_down_grpc_server")
		close(rotationStop)
		grpcServer.GracefulStop()
	}()

	return grpcServer.Serve(lis)
}
