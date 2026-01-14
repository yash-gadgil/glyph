package server

import (
	"context"
	"net"
	"os"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/pkg/telemetry"
	obpb "github.com/yash-gadgil/glyph/services/gen/golang/order_book"
	"github.com/yash-gadgil/glyph/services/mrktdata/handlers"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type grpcServer struct {
	addr string
	log  *zap.Logger
}

func NewGrpcServer(addr string) *grpcServer {
	return &grpcServer{addr: addr, log: logger.New("mrktdata-service")}
}

func (s *grpcServer) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		s.log.Error("failed_to_listen", logger.KV("addr", s.addr), zap.Error(err))
		return err
	}

	telemetry.EnableGRPCHistograms()
	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
		grpc.ChainStreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	)

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcSrv, healthServer)

	handler := handlers.NewMrktdataHandler(ctx)

	if addr := os.Getenv("ORDERBK_SVC_PORT"); addr != "" {
		obConn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			s.log.Warn("orderbook_unavailable_price_sink_disabled", zap.Error(err))
		} else {
			obClient := obpb.NewOrderbookServiceClient(obConn)
			handler.Hub().PriceSink = func(symbol string, priceCents int64) {
				sinkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
				defer cancel()
				if _, err := obClient.InjectPrice(sinkCtx, &obpb.InjectPriceRequest{
					Symbol:     symbol,
					PriceCents: priceCents,
				}); err != nil {
					s.log.Warn("price_sink_inject_failed", logger.KV("symbol", symbol), zap.Error(err))
				}
			}
			go func() {
				<-ctx.Done()
				obConn.Close()
			}()
		}
	} else {
		s.log.Warn("orderbk_svc_port_unset_price_sink_disabled")
	}

	handlers.Register(grpcSrv, handler)

	grpc_prometheus.Register(grpcSrv)
	go telemetry.ServeMetrics(ctx, telemetry.MetricsAddr(), s.log)

	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	s.log.Info("grpc_server_running", logger.KV("addr", s.addr))

	go func() {
		<-ctx.Done()
		healthServer.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
		s.log.Info("shutting_down_grpc_server")
		grpcSrv.GracefulStop()
	}()

	return grpcSrv.Serve(lis)
}
