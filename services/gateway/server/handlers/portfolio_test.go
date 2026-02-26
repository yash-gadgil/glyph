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
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func portfolioConfig(portfolio *mocks.MockPortfolioClient) *handlers.Config {
	return handlers.NewTestConfig(new(mocks.MockAuthClient)).WithPortfolioClient(portfolio)
}

func TestGetPortfolioReturnsBuyingPower(t *testing.T) {
	portfolio := new(mocks.MockPortfolioClient)
	portfolio.On("GetPortfolio", mock.Anything, &userpb.UserSpecifier{UserId: "user-1"}).
		Return(&userpb.PortfolioResponse{
			CashBalanceCents:  4_000_000,
			ReservedCashCents: 1_000_000,
			Currency:          "USD",
			Multiplier:        1,
		}, nil)

	cfg := portfolioConfig(portfolio)
	req := handlers.WithUserID(httptest.NewRequest(http.MethodGet, "/portfolio", nil), "user-1")
	w := httptest.NewRecorder()

	cfg.GetPortfolio(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, float64(3_000_000), body["buying_power_cents"])
	portfolio.AssertExpectations(t)
}

func TestGetHoldingsMapsRows(t *testing.T) {
	portfolio := new(mocks.MockPortfolioClient)
	portfolio.On("GetHoldings", mock.Anything, mock.Anything).
		Return(&userpb.HoldingsResponse{
			Holdings: []*userpb.Holding{{
				Symbol:           "AAPL",
				Qty:              10,
				MarketValueCents: 1_800_000,
			}},
			TotalMarketValueCents: 1_800_000,
		}, nil)

	cfg := portfolioConfig(portfolio)
	req := handlers.WithUserID(httptest.NewRequest(http.MethodGet, "/portfolio/holdings", nil), "user-1")
	w := httptest.NewRecorder()

	cfg.GetHoldings(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Holdings              []map[string]any `json:"holdings"`
		TotalMarketValueCents int64            `json:"total_market_value_cents"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body.Holdings, 1)
	assert.Equal(t, "AAPL", body.Holdings[0]["symbol"])
	portfolio.AssertExpectations(t)
}

func TestGetPortfolioHistoryDefaultsHours(t *testing.T) {
	portfolio := new(mocks.MockPortfolioClient)
	portfolio.On("GetPortfolioHistory", mock.Anything, &userpb.PortfolioHistoryRequest{UserId: "user-1", Hours: 24}).
		Return(&userpb.PortfolioHistoryResponse{
			Points: []*userpb.PortfolioHistoryPoint{{TimeUnix: 1, EquityCents: 100}},
		}, nil)

	cfg := portfolioConfig(portfolio)
	req := handlers.WithUserID(httptest.NewRequest(http.MethodGet, "/portfolio/history", nil), "user-1")
	w := httptest.NewRecorder()

	cfg.GetPortfolioHistory(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	portfolio.AssertExpectations(t)
}

func TestGetPositionsServiceError(t *testing.T) {
	portfolio := new(mocks.MockPortfolioClient)
	portfolio.On("GetPositions", mock.Anything, mock.Anything).
		Return(nil, status.Error(codes.Internal, "down"))

	cfg := portfolioConfig(portfolio)
	req := handlers.WithUserID(httptest.NewRequest(http.MethodGet, "/portfolio/positions", nil), "user-1")
	w := httptest.NewRecorder()

	cfg.GetPositions(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
