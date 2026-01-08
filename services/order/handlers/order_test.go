package handlers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	ordrpb "github.com/yash-gadgil/glyph/services/gen/golang/order"
	obpb "github.com/yash-gadgil/glyph/services/gen/golang/order_book"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"github.com/yash-gadgil/glyph/services/order/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type mockOrderbookClient struct {
	mock.Mock
}

func (m *mockOrderbookClient) AddOrder(ctx context.Context, in *obpb.AddOrderRequest, opts ...grpc.CallOption) (*obpb.AddOrderResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*obpb.AddOrderResponse), args.Error(1)
}

func (m *mockOrderbookClient) CancelOrder(ctx context.Context, in *obpb.CancelOrderRequest, opts ...grpc.CallOption) (*obpb.CancelOrderResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*obpb.CancelOrderResponse), args.Error(1)
}

func (m *mockOrderbookClient) InjectPrice(ctx context.Context, in *obpb.InjectPriceRequest, opts ...grpc.CallOption) (*obpb.InjectPriceResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*obpb.InjectPriceResponse), args.Error(1)
}

type mockAccountClient struct {
	mock.Mock
	userpb.AccountServiceClient
}

func (m *mockAccountClient) ReserveForOrder(ctx context.Context, in *userpb.ReserveForOrderRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *mockAccountClient) ReleaseForOrder(ctx context.Context, in *userpb.ReleaseForOrderRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func newOrderHandler(t *testing.T) (*OrderHandler, sqlmock.Sqlmock, *mockOrderbookClient, *mockAccountClient) {
	t.Helper()
	sdb, dbMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = sdb.Close() })
	obClient := new(mockOrderbookClient)
	userClient := new(mockAccountClient)
	h := NewOrderHandler(sdb, obClient, userClient, nil, zap.NewNop())
	return h, dbMock, obClient, userClient
}

var orderColumns = []string{
	"id", "user_id", "symbol", "side", "order_type", "time_in_force",
	"qty", "filled_qty", "price", "stop_price", "status", "strategy_id",
	"created_at", "updated_at",
}

func orderRow(id, userID uuid.UUID, symbol string, qty, filledQty int64, orderStatus int16) *sqlmock.Rows {
	return sqlmock.NewRows(orderColumns).
		AddRow(id, userID, symbol, int16(0), int16(0), int16(0),
			qty, filledQty, nil, nil, orderStatus, nil,
			time.Now(), time.Now())
}

func TestPlaceOrderValidation(t *testing.T) {
	h, _, _, _ := newOrderHandler(t)

	_, err := h.PlaceOrder(context.Background(), &ordrpb.PlaceOrderRequest{UserId: "bad"})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = h.PlaceOrder(context.Background(), &ordrpb.PlaceOrderRequest{
		UserId: uuid.New().String(), Symbol: "", Qty: 10,
	})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = h.PlaceOrder(context.Background(), &ordrpb.PlaceOrderRequest{
		UserId: uuid.New().String(), Symbol: "AAPL", Qty: 0,
	})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = h.PlaceOrder(context.Background(), &ordrpb.PlaceOrderRequest{
		UserId: uuid.New().String(), Symbol: "AAPL", Qty: 10,
		OrderType: ordrpb.OrderType_LIMIT, Price: 0,
	})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestPlaceOrderMarketBuyNeedsReferencePrice(t *testing.T) {
	h, _, _, _ := newOrderHandler(t)

	_, err := h.PlaceOrder(context.Background(), &ordrpb.PlaceOrderRequest{
		UserId: uuid.New().String(), Symbol: "AAPL", Qty: 10,
		OrderType: ordrpb.OrderType_MARKET, Side: ordrpb.Side_BUY,
	})
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))
}

func TestPlaceOrderReservesThenForwardsToEngine(t *testing.T) {
	h, dbMock, obClient, userClient := newOrderHandler(t)
	userID := uuid.New()
	orderID := uuid.New()

	dbMock.ExpectQuery(`INSERT INTO orders`).
		WillReturnRows(orderRow(orderID, userID, "AAPL", 10, 0, int16(ordrpb.OrderStatus_PENDING)))

	userClient.On("ReserveForOrder", mock.Anything, mock.MatchedBy(func(r *userpb.ReserveForOrderRequest) bool {
		return r.OrderId == orderID.String() && r.CentsPerShare == 5000
	})).Return(&emptypb.Empty{}, nil)

	obClient.On("AddOrder", mock.Anything, mock.MatchedBy(func(r *obpb.AddOrderRequest) bool {
		return r.Id == orderID.String() && r.Symbol == "AAPL"
	})).Return(&obpb.AddOrderResponse{Accepted: true}, nil)

	dbMock.ExpectExec(`UPDATE orders`).WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := h.PlaceOrder(context.Background(), &ordrpb.PlaceOrderRequest{
		UserId: userID.String(), Symbol: "AAPL", Qty: 10,
		OrderType: ordrpb.OrderType_LIMIT, Side: ordrpb.Side_BUY, Price: 5000,
	})
	require.NoError(t, err)
	assert.Equal(t, ordrpb.OrderStatus_OPEN, resp.Order.Status)
	userClient.AssertExpectations(t)
	obClient.AssertExpectations(t)
}

func TestPlaceOrderInsufficientFundsRejects(t *testing.T) {
	h, dbMock, _, userClient := newOrderHandler(t)
	userID := uuid.New()
	orderID := uuid.New()

	dbMock.ExpectQuery(`INSERT INTO orders`).
		WillReturnRows(orderRow(orderID, userID, "AAPL", 10, 0, int16(ordrpb.OrderStatus_PENDING)))
	userClient.On("ReserveForOrder", mock.Anything, mock.Anything).
		Return((*emptypb.Empty)(nil), status.Errorf(codes.FailedPrecondition, "insufficient buying power"))
	dbMock.ExpectExec(`UPDATE orders`).WillReturnResult(sqlmock.NewResult(0, 1))

	_, err := h.PlaceOrder(context.Background(), &ordrpb.PlaceOrderRequest{
		UserId: userID.String(), Symbol: "AAPL", Qty: 10,
		OrderType: ordrpb.OrderType_LIMIT, Side: ordrpb.Side_BUY, Price: 5000,
	})
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))
}

func TestPlaceOrderEngineFailureReleasesReservation(t *testing.T) {
	h, dbMock, obClient, userClient := newOrderHandler(t)
	userID := uuid.New()
	orderID := uuid.New()

	dbMock.ExpectQuery(`INSERT INTO orders`).
		WillReturnRows(orderRow(orderID, userID, "AAPL", 10, 0, int16(ordrpb.OrderStatus_PENDING)))
	userClient.On("ReserveForOrder", mock.Anything, mock.Anything).Return(&emptypb.Empty{}, nil)
	obClient.On("AddOrder", mock.Anything, mock.Anything).
		Return(&obpb.AddOrderResponse{Accepted: false}, fmt.Errorf("engine down"))
	dbMock.ExpectExec(`UPDATE orders`).WillReturnResult(sqlmock.NewResult(0, 1))
	userClient.On("ReleaseForOrder", mock.Anything, mock.MatchedBy(func(r *userpb.ReleaseForOrderRequest) bool {
		return r.OrderId == orderID.String()
	})).Return(&emptypb.Empty{}, nil)

	_, err := h.PlaceOrder(context.Background(), &ordrpb.PlaceOrderRequest{
		UserId: userID.String(), Symbol: "AAPL", Qty: 10,
		OrderType: ordrpb.OrderType_LIMIT, Side: ordrpb.Side_BUY, Price: 5000,
	})
	assert.Equal(t, codes.Internal, status.Code(err))
	userClient.AssertExpectations(t)
}

func TestCancelOrder(t *testing.T) {
	h, dbMock, obClient, userClient := newOrderHandler(t)
	userID := uuid.New()
	orderID := uuid.New()

	dbMock.ExpectQuery(`SELECT .* FROM orders`).
		WithArgs(orderID).
		WillReturnRows(orderRow(orderID, userID, "AAPL", 10, 0, int16(ordrpb.OrderStatus_OPEN)))
	dbMock.ExpectExec(`UPDATE orders`).WithArgs(orderID).WillReturnResult(sqlmock.NewResult(0, 1))
	obClient.On("CancelOrder", mock.Anything, mock.Anything).
		Return(&obpb.CancelOrderResponse{}, nil)
	userClient.On("ReleaseForOrder", mock.Anything, mock.Anything).Return(&emptypb.Empty{}, nil)
	dbMock.ExpectQuery(`SELECT .* FROM orders`).
		WithArgs(orderID).
		WillReturnRows(orderRow(orderID, userID, "AAPL", 10, 0, int16(ordrpb.OrderStatus_CANCELLED)))

	resp, err := h.CancelOrder(context.Background(), &ordrpb.CancelOrderRequest{
		OrderId: orderID.String(),
		UserId:  userID.String(),
	})
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, ordrpb.OrderStatus_CANCELLED, resp.Order.Status)
}

func TestCancelOrderNotFound(t *testing.T) {
	h, dbMock, _, _ := newOrderHandler(t)
	orderID := uuid.New()

	dbMock.ExpectQuery(`SELECT .* FROM orders`).
		WithArgs(orderID).
		WillReturnError(fmt.Errorf("no rows"))

	_, err := h.CancelOrder(context.Background(), &ordrpb.CancelOrderRequest{
		OrderId: orderID.String(),
	})
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestGetOrdersByStatus(t *testing.T) {
	h, dbMock, _, _ := newOrderHandler(t)
	userID := uuid.New()

	dbMock.ExpectQuery(`FROM orders.*AND status`).
		WithArgs(userID, int16(ordrpb.OrderStatus_OPEN), int32(100), int32(0)).
		WillReturnRows(orderRow(uuid.New(), userID, "AAPL", 10, 0, int16(ordrpb.OrderStatus_OPEN)))

	resp, err := h.GetOrders(context.Background(), &ordrpb.GetOrdersRequest{
		UserId: userID.String(),
		Status: ordrpb.OrderStatus_OPEN,
	})
	require.NoError(t, err)
	assert.Len(t, resp.Orders, 1)
}

func TestGetOrdersAllStatuses(t *testing.T) {
	h, dbMock, _, _ := newOrderHandler(t)
	userID := uuid.New()

	dbMock.ExpectQuery(`FROM orders`).
		WithArgs(userID, int32(100), int32(0)).
		WillReturnRows(orderRow(uuid.New(), userID, "AAPL", 10, 0, int16(ordrpb.OrderStatus_FILLED)))

	resp, err := h.GetOrders(context.Background(), &ordrpb.GetOrdersRequest{
		UserId:      userID.String(),
		AllStatuses: true,
	})
	require.NoError(t, err)
	assert.Len(t, resp.Orders, 1)
}

func TestGetOrder(t *testing.T) {
	h, dbMock, _, _ := newOrderHandler(t)
	userID := uuid.New()
	orderID := uuid.New()

	dbMock.ExpectQuery(`SELECT .* FROM orders`).
		WithArgs(orderID).
		WillReturnRows(orderRow(orderID, userID, "AAPL", 10, 3, int16(ordrpb.OrderStatus_PARTIAL_FILL)))

	resp, err := h.GetOrder(context.Background(), &ordrpb.GetOrderRequest{OrderId: orderID.String()})
	require.NoError(t, err)
	assert.Equal(t, orderID.String(), resp.Id)
	assert.Equal(t, int64(3), resp.FilledQty)
}

func TestGetOrderNotFound(t *testing.T) {
	h, dbMock, _, _ := newOrderHandler(t)
	orderID := uuid.New()

	dbMock.ExpectQuery(`SELECT .* FROM orders`).
		WithArgs(orderID).
		WillReturnError(fmt.Errorf("no rows"))

	_, err := h.GetOrder(context.Background(), &ordrpb.GetOrderRequest{OrderId: orderID.String()})
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestPageParams(t *testing.T) {
	l, o := pageParams(0, -5)
	assert.Equal(t, int32(100), l)
	assert.Equal(t, int32(0), o)

	l, _ = pageParams(9999, 0)
	assert.Equal(t, int32(500), l)
}

var fillColumns = []string{
	"trade_id", "order_id", "symbol", "side", "qty", "price_cents", "liquidity", "executed_at",
}

func TestGetFills(t *testing.T) {
	h, dbMock, _, _ := newOrderHandler(t)
	userID := uuid.New()

	dbMock.ExpectQuery(`FROM fills f`).
		WithArgs(userID, int32(100), int32(0)).
		WillReturnRows(sqlmock.NewRows(fillColumns).
			AddRow(uuid.New(), uuid.New(), "AAPL", int16(0), int64(5), int64(5000), "taker", time.Now()))

	resp, err := h.GetFills(context.Background(), &ordrpb.GetFillsRequest{UserId: userID.String()})
	require.NoError(t, err)
	require.Len(t, resp.Fills, 1)
	assert.Equal(t, "AAPL", resp.Fills[0].Symbol)
}

func TestGetStrategyFills(t *testing.T) {
	h, dbMock, _, _ := newOrderHandler(t)
	userID := uuid.New()
	strategyID := uuid.New()

	dbMock.ExpectQuery(`FROM fills f`).
		WithArgs(strategyID, userID, int32(100), int32(0)).
		WillReturnRows(sqlmock.NewRows(fillColumns).
			AddRow(uuid.New(), uuid.New(), "AAPL", int16(1), int64(2), int64(6000), "maker", time.Now()))

	resp, err := h.GetStrategyFills(context.Background(), &ordrpb.GetStrategyFillsRequest{
		StrategyId: strategyID.String(),
		UserId:     userID.String(),
	})
	require.NoError(t, err)
	require.Len(t, resp.Fills, 1)
	assert.Equal(t, ordrpb.Side_SELL, resp.Fills[0].Side)
}

func TestGetStrategyFillsInvalidID(t *testing.T) {
	h, _, _, _ := newOrderHandler(t)

	_, err := h.GetStrategyFills(context.Background(), &ordrpb.GetStrategyFillsRequest{
		StrategyId: "bad",
		UserId:     uuid.New().String(),
	})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestUpdateOrderStatus(t *testing.T) {
	h, dbMock, _, _ := newOrderHandler(t)
	orderID := uuid.New()

	dbMock.ExpectExec(`UPDATE orders`).
		WithArgs(orderID, int64(5), int16(ordrpb.OrderStatus_PARTIAL_FILL)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := h.UpdateOrderStatus(context.Background(), &ordrpb.UpdateOrderStatusRequest{
		OrderId:   orderID.String(),
		Status:    ordrpb.OrderStatus_PARTIAL_FILL,
		FilledQty: 5,
	})
	require.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestApplyFillEventRecordsFill(t *testing.T) {
	h, dbMock, _, _ := newOrderHandler(t)
	tradeID := uuid.New()
	orderID := uuid.New()
	userID := uuid.New()

	dbMock.ExpectBegin()
	dbMock.ExpectExec(`INSERT INTO fills`).WillReturnResult(sqlmock.NewResult(0, 1))
	dbMock.ExpectQuery(`UPDATE orders`).
		WillReturnRows(orderRow(orderID, userID, "AAPL", 10, 5, int16(ordrpb.OrderStatus_PARTIAL_FILL)))
	dbMock.ExpectCommit()

	err := h.ApplyFillEvent(context.Background(), types.FillEvent{
		TradeID: tradeID.String(), OrderID: orderID.String(), Symbol: "AAPL",
		Side: 0, Qty: 5, PriceCents: 5000, Liquidity: "taker", ExecutedAt: time.Now(),
	})
	require.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestApplyFillEventRedeliveryIsNoOp(t *testing.T) {
	h, dbMock, _, _ := newOrderHandler(t)

	dbMock.ExpectBegin()
	dbMock.ExpectExec(`INSERT INTO fills`).WillReturnResult(sqlmock.NewResult(0, 0))
	dbMock.ExpectCommit()

	err := h.ApplyFillEvent(context.Background(), types.FillEvent{
		TradeID: uuid.New().String(), OrderID: uuid.New().String(), Symbol: "AAPL",
		Side: 0, Qty: 5, PriceCents: 5000, Liquidity: "taker", ExecutedAt: time.Now(),
	})
	require.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestApplyFillEventBadID(t *testing.T) {
	h, _, _, _ := newOrderHandler(t)

	err := h.ApplyFillEvent(context.Background(), types.FillEvent{TradeID: "bad"})
	require.Error(t, err)
}

func TestApplyDoneEventFinalizes(t *testing.T) {
	h, dbMock, _, _ := newOrderHandler(t)
	orderID := uuid.New()
	userID := uuid.New()

	dbMock.ExpectQuery(`UPDATE orders`).
		WithArgs(orderID, int16(ordrpb.OrderStatus_FILLED)).
		WillReturnRows(orderRow(orderID, userID, "AAPL", 10, 10, int16(ordrpb.OrderStatus_FILLED)))

	err := h.ApplyDoneEvent(context.Background(), types.DoneEvent{
		OrderID: orderID.String(), Reason: "filled",
	})
	require.NoError(t, err)
}

func TestApplyDoneEventUnknownReason(t *testing.T) {
	h, _, _, _ := newOrderHandler(t)

	err := h.ApplyDoneEvent(context.Background(), types.DoneEvent{
		OrderID: uuid.New().String(), Reason: "huh",
	})
	require.Error(t, err)
}

func TestReservationPrice(t *testing.T) {
	p, err := reservationPrice(&ordrpb.PlaceOrderRequest{Price: 5000})
	require.NoError(t, err)
	assert.Equal(t, int64(5000), p)

	p, err = reservationPrice(&ordrpb.PlaceOrderRequest{Side: ordrpb.Side_BUY, ReferencePriceCents: 1000})
	require.NoError(t, err)
	assert.Equal(t, int64(1050), p)
}

func TestPlaceOrderWithoutUserClientSkipsReservation(t *testing.T) {
	sdb, dbMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer sdb.Close()
	obClient := new(mockOrderbookClient)
	h := NewOrderHandler(sdb, obClient, nil, nil, zap.NewNop())

	userID := uuid.New()
	orderID := uuid.New()
	dbMock.ExpectQuery(`INSERT INTO orders`).
		WillReturnRows(orderRow(orderID, userID, "AAPL", 10, 0, int16(ordrpb.OrderStatus_PENDING)))
	obClient.On("AddOrder", mock.Anything, mock.Anything).
		Return(&obpb.AddOrderResponse{Accepted: true}, nil)
	dbMock.ExpectExec(`UPDATE orders`).WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := h.PlaceOrder(context.Background(), &ordrpb.PlaceOrderRequest{
		UserId: userID.String(), Symbol: "AAPL", Qty: 10,
		OrderType: ordrpb.OrderType_LIMIT, Side: ordrpb.Side_BUY, Price: 5000,
	})
	require.NoError(t, err)
	assert.Equal(t, ordrpb.OrderStatus_OPEN, resp.Order.Status)
	obClient.AssertExpectations(t)
}
