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
	mrktpb "github.com/yash-gadgil/glyph/services/gen/golang/mrktdata"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func exploreConfig(mrkt *mocks.MockMrktdataClient) *handlers.Config {
	return handlers.NewTestConfig(new(mocks.MockAuthClient)).WithMrktdataClient(mrkt)
}

func TestGetNewsReturnsArticles(t *testing.T) {
	mrkt := new(mocks.MockMrktdataClient)
	mrkt.On("GetNews", mock.Anything, mock.Anything).
		Return(&mrktpb.NewsResponse{
			Articles: []*mrktpb.NewsArticle{{Id: "n1", Headline: "Markets rally"}},
		}, nil)

	cfg := exploreConfig(mrkt)
	req := handlers.WithUserID(httptest.NewRequest(http.MethodGet, "/explore/news?symbols=AAPL&limit=5", nil), "user-1")
	w := httptest.NewRecorder()

	cfg.GetNews(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Articles []map[string]any `json:"articles"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body.Articles, 1)
	assert.Equal(t, "Markets rally", body.Articles[0]["headline"])
	mrkt.AssertExpectations(t)
}

func TestGetMoversSplitsGainersLosers(t *testing.T) {
	mrkt := new(mocks.MockMrktdataClient)
	mrkt.On("GetMovers", mock.Anything, mock.Anything).
		Return(&mrktpb.MoversResponse{
			Gainers: []*mrktpb.Mover{{Symbol: "AAPL", ChangePercent: 3.2}},
			Losers:  []*mrktpb.Mover{{Symbol: "TSLA", ChangePercent: -2.1}},
		}, nil)

	cfg := exploreConfig(mrkt)
	req := handlers.WithUserID(httptest.NewRequest(http.MethodGet, "/explore/movers?limit=6", nil), "user-1")
	w := httptest.NewRecorder()

	cfg.GetMovers(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Gainers []map[string]any `json:"gainers"`
		Losers  []map[string]any `json:"losers"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body.Gainers, 1)
	require.Len(t, body.Losers, 1)
	assert.Equal(t, "AAPL", body.Gainers[0]["symbol"])
}

func TestGetNewsServiceError(t *testing.T) {
	mrkt := new(mocks.MockMrktdataClient)
	mrkt.On("GetNews", mock.Anything, mock.Anything).
		Return(nil, status.Error(codes.Internal, "down"))

	cfg := exploreConfig(mrkt)
	req := handlers.WithUserID(httptest.NewRequest(http.MethodGet, "/explore/news", nil), "user-1")
	w := httptest.NewRecorder()

	cfg.GetNews(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
