package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yash-gadgil/glyph/services/gateway/server/handlers"
	"github.com/yash-gadgil/glyph/services/gateway/tests/mocks"
)

func TestConnectToSymbolsRejectsMissingCookie(t *testing.T) {
	cfg := handlers.NewTestConfig(new(mocks.MockAuthClient)).
		WithMrktdataClient(new(mocks.MockMrktdataClient))

	rec := httptest.NewRecorder()
	cfg.ConnectToSymbols(rec, httptest.NewRequest(http.MethodGet, "/watchlists/stream?symbols=AAPL", nil))

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestConnectToWatchlistRejectsMissingCookie(t *testing.T) {
	cfg := handlers.NewTestConfig(new(mocks.MockAuthClient)).
		WithWatchlistClient(new(mocks.MockWatchlistClient)).
		WithMrktdataClient(new(mocks.MockMrktdataClient))

	rec := httptest.NewRecorder()
	cfg.ConnectToWatchlist(rec, httptest.NewRequest(http.MethodGet, "/watchlists/1", nil))

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
