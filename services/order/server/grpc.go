package server

import (
	"context"
	"database/sql"
	"net"
	"os"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/pkg/telemetry"
	mrktpb "github.com/yash-gadgil/glyph/services/gen/golang/mrktdata"
	obpb "github.com/yash-gadgil/glyph/services/gen/golang/order_book"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"github.com/yash-gadgil/glyph/services/order/handlers"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type grpcServer struct {
	addr          string
	db            *sql.DB
	rmqCh         *amqp.Channel
	orderbookAddr string
	log           *zap.Logger
}

func NewGrpcServer(addr string, db *sql.DB, rmqCh *amqp.Channel, orderbookAddr string) *grpcServer {
	return &grpcServer{
		addr:          addr,
		db:            db,
		rmqCh:         rmqCh,
		orderbookAddr: orderbookAddr,
		log:           logger.New("order-service"),
	}
}

func (s *grpcServer) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		s.log.Error("failed_to_listen", logger.KV("addr", s.addr), zap.Error(err))
		return err
	}

	telemetry.EnableGRPCHistograms()
	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logger.UnaryServerInterceptor(s.log),
			grpc_prometheus.UnaryServerInterceptor,
		),
	)

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcSrv, healthServer)

	obConn, err := grpc.NewClient(s.orderbookAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		s.log.Error("orderbook_connect_failed", zap.Error(err))
		return err
	}
	defer obConn.Close()
	obClient := obpb.NewOrderbookServiceClient(obConn)

	var userClient userpb.AccountServiceClient
	if addr := os.Getenv("USER_SVC_PORT"); addr != "" {
		userConn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			s.log.Warn("user_unavailable_reservations_disabled", zap.Error(err))
		} else {
			defer userConn.Close()
			userClient = userpb.NewAccountServiceClient(userConn)
		}
	} else {
		s.log.Warn("user_svc_port_unset_reservations_disabled")
	}

	handler := handlers.NewOrderHandler(s.db, obClient, userClient, s.rmqCh, s.log)
	handlers.Register(grpcSrv, handler)

	grpc_prometheus.Register(grpcSrv)
	go telemetry.ServeMetrics(ctx, telemetry.MetricsAddr(), s.log)

	if addr := os.Getenv("MRKTDATA_SVC_PORT"); addr != "" {
		mrktConn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			s.log.Warn("mrktdata_unavailable_price_feed_disabled", zap.Error(err))
		} else {
			defer mrktConn.Close()
			go handler.RunPriceFeed(ctx, mrktpb.NewMrktdataServiceClient(mrktConn))
		}
	} else {
		s.log.Warn("mrktdata_svc_port_unset_price_feed_disabled")
	}

	go handler.RunDayOrderSweeper(ctx)

	if s.rmqCh != nil {
		if err := ConsumeOrderEvents(ctx, s.rmqCh, handler); err != nil {
			s.log.Warn("order_event_consumer_failed", zap.Error(err))
		}
		if err := StartUserDeletedConsumer(ctx, s.rmqCh, s.db, s.log); err != nil {
			s.log.Warn("user_deleted_consumer_failed", zap.Error(err))
		}
	}

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
