package handlers

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/pkg/telemetry"
	ordrpb "github.com/yash-gadgil/glyph/services/gen/golang/order"
	obpb "github.com/yash-gadgil/glyph/services/gen/golang/order_book"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	db "github.com/yash-gadgil/glyph/services/order/db/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	marketBuySlippageNum = 105
	marketBuySlippageDen = 100
)

func (h *OrderHandler) PlaceOrder(ctx context.Context, req *ordrpb.PlaceOrderRequest) (*ordrpb.PlaceOrderResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	if req.Symbol == "" {
		return nil, status.Errorf(codes.InvalidArgument, "symbol is required")
	}
	if req.Qty <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "quantity must be positive")
	}

	switch req.OrderType {
	case ordrpb.OrderType_LIMIT:
		if req.Price <= 0 {
			return nil, status.Errorf(codes.InvalidArgument, "price required for limit orders")
		}
	case ordrpb.OrderType_STOP:
		if req.StopPrice <= 0 {
			return nil, status.Errorf(codes.InvalidArgument, "stop_price required for stop orders")
		}
	case ordrpb.OrderType_STOP_LIMIT:
		if req.Price <= 0 || req.StopPrice <= 0 {
			return nil, status.Errorf(codes.InvalidArgument, "price and stop_price required for stop-limit orders")
		}
	}

	holdPrice, err := reservationPrice(req)
	if err != nil {
		return nil, err
	}

	displayPrice := req.Price
	if displayPrice <= 0 {
		displayPrice = req.ReferencePriceCents
	}
	if displayPrice <= 0 {
		displayPrice = holdPrice
	}

	var strategyID uuid.NullUUID
	if req.StrategyId != "" {
		parsed, err := uuid.Parse(req.StrategyId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid strategy_id: %v", err)
		}
		strategyID = uuid.NullUUID{UUID: parsed, Valid: true}
	}

	dbOrder, err := h.q.CreateOrder(ctx, db.CreateOrderParams{
		UserID:      userID,
		Symbol:      req.Symbol,
		Side:        int16(req.Side),
		OrderType:   int16(req.OrderType),
		TimeInForce: int16(req.TimeInForce),
		Qty:         req.Qty,
		Price:       sql.NullInt64{Int64: displayPrice, Valid: displayPrice > 0},
		StopPrice:   sql.NullInt64{Int64: req.StopPrice, Valid: req.StopPrice > 0},
		Status:      int16(ordrpb.OrderStatus_PENDING),
		StrategyID:  strategyID,
	})
	if err != nil {
		h.log.Error("persist_order_failed", logger.Action("place_order"), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create order")
	}

	if h.user != nil {
		_, err := h.user.ReserveForOrder(ctx, &userpb.ReserveForOrderRequest{
			OrderId:       dbOrder.ID.String(),
			UserId:        req.UserId,
			Symbol:        req.Symbol,
			Side:          int32(req.Side),
			Qty:           req.Qty,
			CentsPerShare: holdPrice,
		})
		if err != nil {
			h.markRejected(ctx, dbOrder.ID)
			if status.Code(err) == codes.FailedPrecondition {
				return nil, err
			}
			h.log.Error("reservation_failed", logger.KV("order_id", dbOrder.ID.String()), zap.Error(err))
			return nil, status.Errorf(codes.Internal, "unable to verify buying power")
		}
	} else {
		h.log.Warn("reservation_skipped_no_user_client", logger.KV("order_id", dbOrder.ID.String()))
	}

	if req.ReferencePriceCents > 0 && IsMarketOpen(time.Now()) {
		if _, err := h.ob.InjectPrice(ctx, &obpb.InjectPriceRequest{
			Symbol:     req.Symbol,
			PriceCents: req.ReferencePriceCents,
		}); err != nil {
			h.log.Warn("price_priming_failed", logger.KV("symbol", req.Symbol), zap.Error(err))
		}
	}

	obResp, err := h.ob.AddOrder(ctx, &obpb.AddOrderRequest{
		Id:          dbOrder.ID.String(),
		UserId:      req.UserId,
		Symbol:      req.Symbol,
		Side:        obpb.Side(req.Side),
		OrderType:   obpb.OrderType(req.OrderType),
		TimeInForce: obpb.TimeInForce(req.TimeInForce),
		Qty:         req.Qty,
		Price:       req.Price,
		StopPrice:   req.StopPrice,
	})
	if err != nil || !obResp.Accepted {
		h.log.Error("orderbook_rejected", logger.KV("order_id", dbOrder.ID.String()), zap.Error(err))
		h.markRejected(ctx, dbOrder.ID)
		h.releaseReservation(ctx, dbOrder.ID)
		return nil, status.Errorf(codes.Internal, "orderbook rejected order")
	}

	if err := h.q.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		ID:     dbOrder.ID,
		Status: int16(ordrpb.OrderStatus_OPEN),
	}); err != nil {
		h.log.Error("mark_open_failed", logger.KV("order_id", dbOrder.ID.String()), zap.Error(err))
	}
	dbOrder.Status = int16(ordrpb.OrderStatus_OPEN)

	telemetry.OrdersPlacedTotal.WithLabelValues(telemetry.SideLabel(int32(req.Side))).Inc()

	return &ordrpb.PlaceOrderResponse{
		Order: dbOrderToProto(dbOrder),
	}, nil
}

func (h *OrderHandler) CancelOrder(ctx context.Context, req *ordrpb.CancelOrderRequest) (*ordrpb.CancelOrderResponse, error) {
	orderID, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid order_id")
	}

	order, err := h.q.GetOrderById(ctx, orderID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "order not found")
	}

	if err := h.q.CancelOrder(ctx, orderID); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to cancel order")
	}

	if _, err := h.ob.CancelOrder(ctx, &obpb.CancelOrderRequest{
		OrderId: req.OrderId,
		Symbol:  order.Symbol,
	}); err != nil {
		h.log.Warn("orderbook_cancel_failed", logger.KV("order_id", orderID.String()), zap.Error(err))
	}

	h.releaseReservation(ctx, orderID)

	updated, err := h.q.GetOrderById(ctx, orderID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch updated order")
	}

	return &ordrpb.CancelOrderResponse{
		Success: true,
		Order:   dbOrderToProto(updated),
	}, nil
}

func reservationPrice(req *ordrpb.PlaceOrderRequest) (int64, error) {
	if req.Price > 0 {
		return req.Price, nil
	}

	if req.Side == ordrpb.Side_BUY {
		if req.ReferencePriceCents <= 0 {
			return 0, status.Errorf(codes.FailedPrecondition, "reference price unavailable for market buy")
		}
		return req.ReferencePriceCents * marketBuySlippageNum / marketBuySlippageDen, nil
	}

	if req.ReferencePriceCents > 0 {
		return req.ReferencePriceCents, nil
	}
	if req.StopPrice > 0 {
		return req.StopPrice, nil
	}
	return 1, nil
}

func (h *OrderHandler) markRejected(ctx context.Context, orderID uuid.UUID) {
	if err := h.q.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		ID:     orderID,
		Status: int16(ordrpb.OrderStatus_REJECTED),
	}); err != nil {
		h.log.Error("mark_rejected_failed", logger.KV("order_id", orderID.String()), zap.Error(err))
	}
}

func (h *OrderHandler) releaseReservation(ctx context.Context, orderID uuid.UUID) {
	if h.user == nil {
		return
	}
	if _, err := h.user.ReleaseForOrder(ctx, &userpb.ReleaseForOrderRequest{
		OrderId: orderID.String(),
	}); err != nil {
		h.log.Error("release_reservation_failed", logger.KV("order_id", orderID.String()), zap.Error(err))
	}
}

func dbOrderToProto(o db.Order) *ordrpb.Order {
	po := &ordrpb.Order{
		Id:          o.ID.String(),
		UserId:      o.UserID.String(),
		Symbol:      o.Symbol,
		Side:        ordrpb.Side(o.Side),
		OrderType:   ordrpb.OrderType(o.OrderType),
		TimeInForce: ordrpb.TimeInForce(o.TimeInForce),
		Qty:         o.Qty,
		FilledQty:   o.FilledQty,
		Status:      ordrpb.OrderStatus(o.Status),
		CreatedAt:   timestamppb.New(o.CreatedAt),
		UpdatedAt:   timestamppb.New(o.UpdatedAt),
	}
	if o.Price.Valid {
		po.Price = o.Price.Int64
	}
	if o.StopPrice.Valid {
		po.StopPrice = o.StopPrice.Int64
	}
	return po
}
