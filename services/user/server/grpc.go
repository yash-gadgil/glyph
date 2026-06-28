package server

import (
	"context"
	"database/sql"
	"net"
	"os"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/pkg/telemetry"
	mrktpb "github.com/yash-gadgil/glyph/services/gen/golang/mrktdata"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"github.com/yash-gadgil/glyph/services/user/handlers"
	"github.com/yash-gadgil/glyph/services/user/worker"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type gRPCServer struct {
	db   *sql.DB
	addr string
	pub  handlers.UserEventPublisher
	log  *zap.Logger
}

func NewGrpcServer(addr string, db *sql.DB, pub handlers.UserEventPublisher) *gRPCServer {
	return &gRPCServer{db: db, addr: addr, pub: pub, log: logger.New("user-service")}
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

	var prices mrktpb.MrktdataServiceClient
	if addr := os.Getenv("MRKTDATA_SVC_PORT"); addr != "" {
		mrktConn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			s.log.Warn("mrktdata_unavailable_holdings_valued_at_cost", zap.Error(err))
		} else {
			defer mrktConn.Close()
			prices = mrktpb.NewMrktdataServiceClient(mrktConn)
		}
	} else {
		s.log.Warn("mrktdata_svc_port_unset_holdings_valued_at_cost")
	}

	userpb.RegisterWatchlistServiceServer(grpcServer, handlers.NewWatchlistHandler(s.db, s.log))
	userpb.RegisterAccountServiceServer(grpcServer, handlers.NewAccountHandler(s.db, s.pub, s.log))
	userpb.RegisterPortfolioServiceServer(grpcServer, handlers.NewPortfolioHandler(s.db, prices, s.log))

	grpc_prometheus.Register(grpcServer)
	go telemetry.ServeMetrics(ctx, telemetry.MetricsAddr(), s.log)

	go worker.NewSnapshotter(s.db, prices, s.log).Run(ctx)

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
