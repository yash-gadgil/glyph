package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/pkg/telemetry"
	ordrpb "github.com/yash-gadgil/glyph/services/gen/golang/order"
	obpb "github.com/yash-gadgil/glyph/services/gen/golang/order_book"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	db "github.com/yash-gadgil/glyph/services/order/db/gen"
	"github.com/yash-gadgil/glyph/services/order/types"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	defaultPageLimit = 100
	maxPageLimit     = 500

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

func (h *OrderHandler) GetOrders(ctx context.Context, req *ordrpb.GetOrdersRequest) (*ordrpb.GetOrdersResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id")
	}

	limit, offset := pageParams(req.Limit, req.Offset)

	var orders []db.Order
	if req.AllStatuses {
		orders, err = h.q.GetOrdersByUser(ctx, db.GetOrdersByUserParams{
			UserID: userID,
			Limit:  limit,
			Offset: offset,
		})
	} else {
		orders, err = h.q.GetOrdersByUserAndStatus(ctx, db.GetOrdersByUserAndStatusParams{
			UserID: userID,
			Status: int16(req.Status),
			Limit:  limit,
			Offset: offset,
		})
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch orders")
	}

	protoOrders := make([]*ordrpb.Order, len(orders))
	for i, o := range orders {
		protoOrders[i] = dbOrderToProto(o)
	}

	return &ordrpb.GetOrdersResponse{Orders: protoOrders}, nil
}

func (h *OrderHandler) GetOrder(ctx context.Context, req *ordrpb.GetOrderRequest) (*ordrpb.Order, error) {
	orderID, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid order_id")
	}

	order, err := h.q.GetOrderById(ctx, orderID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "order not found")
	}

	return dbOrderToProto(order), nil
}

func (h *OrderHandler) GetFills(ctx context.Context, req *ordrpb.GetFillsRequest) (*ordrpb.GetFillsResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id")
	}

	limit, offset := pageParams(req.Limit, req.Offset)

	rows, err := h.q.GetFillsByUser(ctx, db.GetFillsByUserParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch fills")
	}

	fills := make([]*ordrpb.Fill, len(rows))
	for i, f := range rows {
		fills[i] = &ordrpb.Fill{
			TradeId:    f.TradeID.String(),
			OrderId:    f.OrderID.String(),
			Symbol:     f.Symbol,
			Side:       ordrpb.Side(f.Side),
			Qty:        f.Qty,
			PriceCents: f.PriceCents,
			Liquidity:  f.Liquidity,
			ExecutedAt: timestamppb.New(f.ExecutedAt),
		}
	}

	return &ordrpb.GetFillsResponse{Fills: fills}, nil
}

func (h *OrderHandler) GetStrategyFills(ctx context.Context, req *ordrpb.GetStrategyFillsRequest) (*ordrpb.GetFillsResponse, error) {
	strategyID, err := uuid.Parse(req.StrategyId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid strategy_id")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id")
	}

	limit, offset := pageParams(req.Limit, req.Offset)

	rows, err := h.q.GetStrategyFills(ctx, db.GetStrategyFillsParams{
		StrategyID: uuid.NullUUID{UUID: strategyID, Valid: true},
		UserID:     userID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch strategy fills")
	}

	fills := make([]*ordrpb.Fill, len(rows))
	for i, f := range rows {
		fills[i] = &ordrpb.Fill{
			TradeId:    f.TradeID.String(),
			OrderId:    f.OrderID.String(),
			Symbol:     f.Symbol,
			Side:       ordrpb.Side(f.Side),
			Qty:        f.Qty,
			PriceCents: f.PriceCents,
			Liquidity:  f.Liquidity,
			ExecutedAt: timestamppb.New(f.ExecutedAt),
		}
	}

	return &ordrpb.GetFillsResponse{Fills: fills}, nil
}

func (h *OrderHandler) UpdateOrderStatus(ctx context.Context, req *ordrpb.UpdateOrderStatusRequest) (*ordrpb.UpdateOrderStatusResponse, error) {
	orderID, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid order_id")
	}

	if err := h.q.UpdateOrderFill(ctx, db.UpdateOrderFillParams{
		ID:        orderID,
		FilledQty: req.FilledQty,
		Status:    int16(req.Status),
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update order status")
	}

	return &ordrpb.UpdateOrderStatusResponse{Success: true}, nil
}

func (h *OrderHandler) ApplyFillEvent(ctx context.Context, event types.FillEvent) error {
	tradeID, err := uuid.Parse(event.TradeID)
	if err != nil {
		return fmt.Errorf("bad trade_id %q: %w", event.TradeID, err)
	}
	orderID, err := uuid.Parse(event.OrderID)
	if err != nil {
		return fmt.Errorf("bad order_id %q: %w", event.OrderID, err)
	}
	if event.Qty <= 0 {
		return fmt.Errorf("invalid fill qty %d", event.Qty)
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback()

	qtx := h.q.WithTx(tx)

	inserted, err := qtx.InsertFill(ctx, db.InsertFillParams{
		TradeID:    tradeID,
		OrderID:    orderID,
		Symbol:     event.Symbol,
		Side:       event.Side,
		Qty:        event.Qty,
		PriceCents: event.PriceCents,
		Liquidity:  event.Liquidity,
		ExecutedAt: event.ExecutedAt,
	})
	if err != nil {
		return fmt.Errorf("insert fill: %w", err)
	}
	if inserted == 0 {
		return tx.Commit()
	}

	if _, err := qtx.ApplyFillToOrder(ctx, db.ApplyFillToOrderParams{
		ID:        orderID,
		FilledQty: event.Qty,
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("fill for unknown order %s: %w", orderID, err)
		}
		return fmt.Errorf("apply fill: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

func (h *OrderHandler) ApplyDoneEvent(ctx context.Context, event types.DoneEvent) error {
	orderID, err := uuid.Parse(event.OrderID)
	if err != nil {
		return fmt.Errorf("bad order_id %q: %w", event.OrderID, err)
	}

	var terminal ordrpb.OrderStatus
	switch event.Reason {
	case "filled":
		terminal = ordrpb.OrderStatus_FILLED
	case "cancelled", "ioc_expired", "fok_killed":
		terminal = ordrpb.OrderStatus_CANCELLED
	default:
		return fmt.Errorf("unknown done reason %q", event.Reason)
	}

	if _, err := h.q.FinalizeOrder(ctx, db.FinalizeOrderParams{
		ID:     orderID,
		Status: int16(terminal),
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("finalize order: %w", err)
	}

	return nil
}

func pageParams(limit, offset int32) (int32, int32) {
	if limit <= 0 {
		limit = defaultPageLimit
	}
	if limit > maxPageLimit {
		limit = maxPageLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
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
