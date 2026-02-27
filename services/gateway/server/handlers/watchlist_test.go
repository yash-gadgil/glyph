package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yash-gadgil/glyph/services/gateway/server/handlers"
	"github.com/yash-gadgil/glyph/services/gateway/tests/mocks"
	mrktpb "github.com/yash-gadgil/glyph/services/gen/golang/mrktdata"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func watchlistConfig(watchlist *mocks.MockWatchlistClient, mrkt *mocks.MockMrktdataClient) *handlers.Config {
	return handlers.NewTestConfig(new(mocks.MockAuthClient)).
		WithWatchlistClient(watchlist).
		WithMrktdataClient(mrkt)
}

func TestGetWatchlistsSuccess(t *testing.T) {
	watchlist := new(mocks.MockWatchlistClient)
	watchlist.On("GetWatchlists", mock.Anything, &userpb.UserSpecifier{UserId: "user-1"}).
		Return(&userpb.WatchlistsResponse{}, nil)

	cfg := watchlistConfig(watchlist, new(mocks.MockMrktdataClient))
	req := handlers.WithUserID(httptest.NewRequest(http.MethodGet, "/watchlists", nil), "user-1")
	w := httptest.NewRecorder()

	cfg.GetWatchlists(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	watchlist.AssertExpectations(t)
}

func TestCreateWatchlistConflict(t *testing.T) {
	watchlist := new(mocks.MockWatchlistClient)
	watchlist.On("CreateWatchlist", mock.Anything, mock.Anything).
		Return(nil, status.Error(codes.AlreadyExists, "name taken"))

	cfg := watchlistConfig(watchlist, new(mocks.MockMrktdataClient))

	body, _ := json.Marshal(map[string]string{"name": "Tech"})
	req := handlers.WithUserID(httptest.NewRequest(http.MethodPost, "/watchlists", bytes.NewReader(body)), "user-1")
	w := httptest.NewRecorder()

	cfg.CreateWatchlist(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestModifyWatchlistRejectsUnknownSymbol(t *testing.T) {
	watchlist := new(mocks.MockWatchlistClient)
	mrkt := new(mocks.MockMrktdataClient)
	mrkt.On("GetAvailableSymbols", mock.Anything, mock.Anything).
		Return(&mrktpb.AvailableSymbolsResponse{
			Symbols: []*mrktpb.Symbol{{Name: "AAPL"}},
		}, nil)

	cfg := watchlistConfig(watchlist, mrkt)

	body, _ := json.Marshal(map[string][]string{"symbols": {"DOGE"}})
	req := handlers.WithUserID(httptest.NewRequest(http.MethodPatch, "/watchlists/1?action=subscribe", bytes.NewReader(body)), "user-1")
	w := httptest.NewRecorder()

	cfg.ModifyWatchlist(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mrkt.AssertExpectations(t)
}

func TestGetAvailableSymbolsRanksMatches(t *testing.T) {
	mrkt := new(mocks.MockMrktdataClient)
	mrkt.On("GetAvailableSymbols", mock.Anything, mock.Anything).
		Return(&mrktpb.AvailableSymbolsResponse{
			Symbols: []*mrktpb.Symbol{
				{Name: "AAPL", CompanyName: "Apple Inc"},
				{Name: "TSLA", CompanyName: "Tesla Inc"},
			},
		}, nil)

	cfg := watchlistConfig(new(mocks.MockWatchlistClient), mrkt)
	req := handlers.WithUserID(httptest.NewRequest(http.MethodGet, "/watchlists/symbols?q=AAPL", nil), "user-1")
	w := httptest.NewRecorder()

	cfg.GetAvailableSymbols(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body mrktpb.AvailableSymbolsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body.Symbols, 1)
	assert.Equal(t, "AAPL", body.Symbols[0].Name)
}

func TestDeleteWatchlistNoContent(t *testing.T) {
	watchlist := new(mocks.MockWatchlistClient)
	watchlist.On("DeleteWatchlist", mock.Anything, mock.Anything).
		Return(&emptypb.Empty{}, nil)

	cfg := watchlistConfig(watchlist, new(mocks.MockMrktdataClient))
	req := handlers.WithUserID(httptest.NewRequest(http.MethodDelete, "/watchlists/1", nil), "user-1")
	w := httptest.NewRecorder()

	cfg.DeleteWatchlist(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	watchlist.AssertExpectations(t)
}
