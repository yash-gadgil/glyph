package server

import (
	"context"
	"net"
	"os"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/pkg/telemetry"
	"github.com/yash-gadgil/glyph/services/advisor/cache"
	"github.com/yash-gadgil/glyph/services/advisor/handlers"
	"github.com/yash-gadgil/glyph/services/advisor/llm"
	"github.com/yash-gadgil/glyph/services/advisor/types"
	advisorpb "github.com/yash-gadgil/glyph/services/gen/golang/advisor"
	inferpb "github.com/yash-gadgil/glyph/services/gen/golang/inference"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type gRPCServer struct {
	addr string
	log  *zap.Logger
}

func NewGrpcServer(addr string) *gRPCServer {
	return &gRPCServer{addr: addr, log: logger.New("advisor-service")}
}

func geminiModel() string {
	if m := os.Getenv("GEMINI_MODEL"); m != "" {
		return m
	}
	return "gemini-2.5-flash"
}

func (s *gRPCServer) Run(ctx context.Context) error {
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

	var portfolio userpb.PortfolioServiceClient
	if addr := os.Getenv("USER_SVC_PORT"); addr != "" {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			s.log.Warn("user_service_unavailable", zap.Error(err))
		} else {
			defer conn.Close()
			portfolio = userpb.NewPortfolioServiceClient(conn)
		}
	} else {
		s.log.Warn("user_svc_port_unset")
	}

	var model types.Provider
	switch os.Getenv("LLM_PROVIDER") {
	case "inference":
		if addr := os.Getenv("INFERENCE_SVC_PORT"); addr != "" {
			conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				s.log.Warn("inference_service_unavailable", zap.Error(err))
			} else {
				defer conn.Close()
				model = llm.NewInference(inferpb.NewInferenceServiceClient(conn))
			}
		} else {
			s.log.Warn("inference_svc_port_unset")
		}
	default:
		if key := os.Getenv("GEMINI_API_KEY"); key != "" {
			provider, err := llm.NewGemini(ctx, key, geminiModel())
			if err != nil {
				s.log.Warn("gemini_unavailable", zap.Error(err))
			} else {
				model = provider
			}
		} else {
			s.log.Warn("gemini_api_key_unset")
		}
	}

	analysisCache := cache.Init(ctx, s.log)

	advisorpb.RegisterAdvisorServiceServer(grpcServer, handlers.NewAdvisorHandler(portfolio, model, analysisCache, s.log))

	grpc_prometheus.Register(grpcServer)
	go telemetry.ServeMetrics(ctx, telemetry.MetricsAddr(), s.log)

	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	s.log.Info("grpc_server_running", logger.Action("running_grpc_server"), logger.KV("addr", s.addr))

	go func() {
		<-ctx.Done()
		healthServer.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
		s.log.Info("shutting_down_grpc_server")
		grpcServer.GracefulStop()
	}()

	return grpcServer.Serve(lis)
}
