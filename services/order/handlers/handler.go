package handlers

import (
	"database/sql"

	amqp "github.com/rabbitmq/amqp091-go"
	ordrpb "github.com/yash-gadgil/glyph/services/gen/golang/order"
	obpb "github.com/yash-gadgil/glyph/services/gen/golang/order_book"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	db "github.com/yash-gadgil/glyph/services/order/db/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type OrderHandler struct {
	ordrpb.UnimplementedOrderServiceServer
	db    *sql.DB
	q     *db.Queries
	ob    obpb.OrderbookServiceClient
	user  userpb.AccountServiceClient
	rmqCh *amqp.Channel
	log   *zap.Logger
}

func NewOrderHandler(sdb *sql.DB, ob obpb.OrderbookServiceClient, user userpb.AccountServiceClient, rmqCh *amqp.Channel, log *zap.Logger) *OrderHandler {
	return &OrderHandler{
		db:    sdb,
		q:     db.New(sdb),
		ob:    ob,
		user:  user,
		rmqCh: rmqCh,
		log:   log,
	}
}

func Register(grpcServer *grpc.Server, h *OrderHandler) {
	ordrpb.RegisterOrderServiceServer(grpcServer, h)
}
