package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"context"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yash-gadgil/glyph/services/gateway/server/handlers"
	"github.com/yash-gadgil/glyph/services/gateway/tests/mocks"
	mrktpb "github.com/yash-gadgil/glyph/services/gen/golang/mrktdata"
	ordrpb "github.com/yash-gadgil/glyph/services/gen/golang/order"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func orderConfig(order *mocks.MockOrderClient, mrkt *mocks.MockMrktdataClient) *handlers.Config {
	return handlers.NewTestConfig(new(mocks.MockAuthClient)).
		WithOrderClient(order).
		WithMrktdataClient(mrkt)
}

func TestCreateOrderPlacesWithReferencePrice(t *testing.T) {
	order := new(mocks.MockOrderClient)
	mrkt := new(mocks.MockMrktdataClient)

	mrkt.On("GetLatestPrices", mock.Anything, mock.Anything).
		Return(&mrktpb.LatestPricesResponse{
			Prices: []*mrktpb.SymbolPrice{{Symbol: "AAPL", PriceCents: 18_000}},
		}, nil)
	order.On("PlaceOrder", mock.Anything, mock.MatchedBy(func(req *ordrpb.PlaceOrderRequest) bool {
		return req.Side == ordrpb.Side_BUY && req.ReferencePriceCents == 18_000
	})).Return(&ordrpb.PlaceOrderResponse{Order: &ordrpb.Order{Id: "o1", Symbol: "AAPL"}}, nil)

	cfg := orderConfig(order, mrkt)
	body, _ := json.Marshal(map[string]any{
		"symbol": "AAPL", "side": "buy", "orderType": "market", "timeInForce": "day", "quantity": 5,
	})
	req := handlers.WithUserID(httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body)), "user-1")
	w := httptest.NewRecorder()

	cfg.CreateOrder(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	order.AssertExpectations(t)
	mrkt.AssertExpectations(t)
}

func TestCreateOrderRejectsBadSide(t *testing.T) {
	cfg := orderConfig(new(mocks.MockOrderClient), new(mocks.MockMrktdataClient))

	body, _ := json.Marshal(map[string]any{
		"symbol": "AAPL", "side": "hodl", "orderType": "market", "timeInForce": "day", "quantity": 5,
	})
	req := handlers.WithUserID(httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body)), "user-1")
	w := httptest.NewRecorder()

	cfg.CreateOrder(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateOrderMapsFailedPrecondition(t *testing.T) {
	order := new(mocks.MockOrderClient)
	mrkt := new(mocks.MockMrktdataClient)
	mrkt.On("GetLatestPrices", mock.Anything, mock.Anything).Return(&mrktpb.LatestPricesResponse{}, nil)
	order.On("PlaceOrder", mock.Anything, mock.Anything).
		Return(nil, status.Error(codes.FailedPrecondition, "insufficient buying power"))

	cfg := orderConfig(order, mrkt)
	body, _ := json.Marshal(map[string]any{
		"symbol": "AAPL", "side": "buy", "orderType": "market", "timeInForce": "day", "quantity": 5,
	})
	req := handlers.WithUserID(httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body)), "user-1")
	w := httptest.NewRecorder()

	cfg.CreateOrder(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestGetOrdersReturnsMapped(t *testing.T) {
	order := new(mocks.MockOrderClient)
	order.On("GetOrders", mock.Anything, mock.MatchedBy(func(req *ordrpb.GetOrdersRequest) bool {
		return req.AllStatuses
	})).Return(&ordrpb.GetOrdersResponse{
		Orders: []*ordrpb.Order{{Id: "o1", Symbol: "AAPL", Side: ordrpb.Side_BUY, Status: ordrpb.OrderStatus_OPEN}},
	}, nil)

	cfg := orderConfig(order, new(mocks.MockMrktdataClient))
	req := handlers.WithUserID(httptest.NewRequest(http.MethodGet, "/orders", nil), "user-1")
	w := httptest.NewRecorder()

	cfg.GetOrders(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Orders []map[string]any `json:"orders"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body.Orders, 1)
	assert.Equal(t, "buy", body.Orders[0]["side"])
	assert.Equal(t, "open", body.Orders[0]["status"])
	order.AssertExpectations(t)
}

func TestDeleteOrderSuccess(t *testing.T) {
	order := new(mocks.MockOrderClient)
	order.On("CancelOrder", mock.Anything, mock.Anything).
		Return(&ordrpb.CancelOrderResponse{Success: true, Order: &ordrpb.Order{Id: "o1"}}, nil)

	cfg := orderConfig(order, new(mocks.MockMrktdataClient))
	req := handlers.WithUserID(httptest.NewRequest(http.MethodDelete, "/orders/o1", nil), "user-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "o1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	cfg.DeleteOrder(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	order.AssertExpectations(t)
}
