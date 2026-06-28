package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yash-gadgil/glyph/services/gateway/server/handlers"
	"github.com/yash-gadgil/glyph/services/gateway/tests/mocks"
	ordrpb "github.com/yash-gadgil/glyph/services/gen/golang/order"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func accountConfig(account *mocks.MockAccountClient, portfolio *mocks.MockPortfolioClient, order *mocks.MockOrderClient) *handlers.Config {
	return handlers.NewTestConfig(new(mocks.MockAuthClient)).
		WithAccountClient(account).
		WithPortfolioClient(portfolio).
		WithOrderClient(order)
}

func TestMeReturnsUserID(t *testing.T) {
	cfg := accountConfig(new(mocks.MockAccountClient), new(mocks.MockPortfolioClient), new(mocks.MockOrderClient))

	req := handlers.WithUserID(httptest.NewRequest(http.MethodGet, "/account/me", nil), "user-42")
	w := httptest.NewRecorder()

	cfg.Me(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Id string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "user-42", body.Id)
}

func TestGetAccountComposesProfileBalancesAndEquity(t *testing.T) {
	account := new(mocks.MockAccountClient)
	portfolio := new(mocks.MockPortfolioClient)
	cfg := accountConfig(account, portfolio, new(mocks.MockOrderClient))

	account.On("GetProfile", mock.Anything, &userpb.UserSpecifier{UserId: "user-1"}).
		Return(&userpb.Profile{UserId: "user-1", Email: "u@example.com", UserName: "Yash"}, nil)
	portfolio.On("GetPortfolio", mock.Anything, mock.Anything).
		Return(&userpb.PortfolioResponse{
			CashBalanceCents:  10_000_000,
			ReservedCashCents: 500_000,
			Currency:          "USD",
			Multiplier:        1,
		}, nil)
	portfolio.On("GetHoldings", mock.Anything, mock.Anything).
		Return(&userpb.HoldingsResponse{TotalMarketValueCents: 2_000_000}, nil)

	req := handlers.WithUserID(httptest.NewRequest(http.MethodGet, "/account", nil), "user-1")
	w := httptest.NewRecorder()

	cfg.GetAccount(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "u@example.com", body["email"])
	assert.Equal(t, float64(9_500_000), body["buying_power_cents"])
	assert.Equal(t, float64(12_000_000), body["equity_cents"])
	account.AssertExpectations(t)
	portfolio.AssertExpectations(t)
}

func TestGetAccountMissingUser(t *testing.T) {
	cfg := accountConfig(new(mocks.MockAccountClient), new(mocks.MockPortfolioClient), new(mocks.MockOrderClient))

	w := httptest.NewRecorder()
	cfg.GetAccount(w, httptest.NewRequest(http.MethodGet, "/account", nil))

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestResetAccountSuccess(t *testing.T) {
	account := new(mocks.MockAccountClient)
	cfg := accountConfig(account, new(mocks.MockPortfolioClient), new(mocks.MockOrderClient))

	account.On("ResetAccount", mock.Anything, &userpb.UserSpecifier{UserId: "user-1"}).
		Return(&emptypb.Empty{}, nil)

	req := handlers.WithUserID(httptest.NewRequest(http.MethodPost, "/account/reset", nil), "user-1")
	w := httptest.NewRecorder()

	cfg.ResetAccount(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	account.AssertExpectations(t)
}

func TestDeleteAccountSuccess(t *testing.T) {
	account := new(mocks.MockAccountClient)
	cfg := accountConfig(account, new(mocks.MockPortfolioClient), new(mocks.MockOrderClient))

	account.On("DeleteAccount", mock.Anything, &userpb.UserSpecifier{UserId: "user-1"}).
		Return(&emptypb.Empty{}, nil)

	req := handlers.WithUserID(httptest.NewRequest(http.MethodDelete, "/account", nil), "user-1")
	w := httptest.NewRecorder()

	cfg.DeleteAccount(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	account.AssertExpectations(t)
}

func TestDeleteAccountMissingUser(t *testing.T) {
	cfg := accountConfig(new(mocks.MockAccountClient), new(mocks.MockPortfolioClient), new(mocks.MockOrderClient))

	w := httptest.NewRecorder()
	cfg.DeleteAccount(w, httptest.NewRequest(http.MethodDelete, "/account", nil))

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDeleteAccountServiceError(t *testing.T) {
	account := new(mocks.MockAccountClient)
	cfg := accountConfig(account, new(mocks.MockPortfolioClient), new(mocks.MockOrderClient))

	account.On("DeleteAccount", mock.Anything, &userpb.UserSpecifier{UserId: "user-1"}).
		Return(nil, status.Error(codes.NotFound, "user not found"))

	req := handlers.WithUserID(httptest.NewRequest(http.MethodDelete, "/account", nil), "user-1")
	w := httptest.NewRecorder()

	cfg.DeleteAccount(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	account.AssertExpectations(t)
}

func TestGetProfileServiceError(t *testing.T) {
	account := new(mocks.MockAccountClient)
	cfg := accountConfig(account, new(mocks.MockPortfolioClient), new(mocks.MockOrderClient))

	account.On("GetProfile", mock.Anything, mock.Anything).
		Return(nil, status.Error(codes.NotFound, "no user"))

	req := handlers.WithUserID(httptest.NewRequest(http.MethodGet, "/account/profile", nil), "user-1")
	w := httptest.NewRecorder()

	cfg.GetProfile(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetTradesReturnsFills(t *testing.T) {
	order := new(mocks.MockOrderClient)
	cfg := accountConfig(new(mocks.MockAccountClient), new(mocks.MockPortfolioClient), order)

	order.On("GetFills", mock.Anything, &ordrpb.GetFillsRequest{
		UserId: "user-1",
		Limit:  10,
	}).Return(&ordrpb.GetFillsResponse{
		Fills: []*ordrpb.Fill{{
			TradeId:    "t1",
			OrderId:    "o1",
			Symbol:     "AAPL",
			Side:       ordrpb.Side_BUY,
			Qty:        5,
			PriceCents: 18_000,
			Liquidity:  "taker",
			ExecutedAt: timestamppb.Now(),
		}},
	}, nil)

	req := handlers.WithUserID(httptest.NewRequest(http.MethodGet, "/account/trades?limit=10", nil), "user-1")
	w := httptest.NewRecorder()

	cfg.GetTrades(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Fills []map[string]any `json:"fills"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body.Fills, 1)
	assert.Equal(t, "AAPL", body.Fills[0]["symbol"])
	assert.Equal(t, "buy", body.Fills[0]["side"])
	assert.NotEmpty(t, body.Fills[0]["executed_at"])
	order.AssertExpectations(t)
}
