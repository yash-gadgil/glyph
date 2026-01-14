package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	alpaca "github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	mrktdata "github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mrktpb "github.com/yash-gadgil/glyph/services/gen/golang/mrktdata"
	"google.golang.org/protobuf/types/known/emptypb"
)

func assetsClient(url string) *alpaca.Client {
	return alpaca.NewClient(alpaca.ClientOpts{BaseURL: url, RetryLimit: 1})
}

func fakeBarsServer(t *testing.T, capture *[]*http.Request) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if capture != nil {
			*capture = append(*capture, r.Clone(context.Background()))
		}
		bars := map[string]any{}
		for _, symbol := range strings.Split(r.URL.Query().Get("symbols"), ",") {
			if symbol == "" {
				continue
			}
			bars[symbol] = []map[string]any{
				{"t": "2026-01-02T15:00:00Z", "o": 100.5, "h": 101.0, "l": 99.5, "c": 100.75, "v": 1000, "n": 10, "vw": 100.6},
				{"t": "2026-01-02T16:00:00Z", "o": 100.75, "h": 102.0, "l": 100.5, "c": 101.5, "v": 2000, "n": 20, "vw": 101.2},
			}
		}
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"next_page_token": nil,
			"bars":            bars,
		}))
	}))
}

func barsClient(url string) *mrktdata.Client {
	return mrktdata.NewClient(mrktdata.ClientOpts{BaseURL: url, RetryLimit: 1})
}

func TestGetHistoricalStockDataMapsBars(t *testing.T) {
	var reqs []*http.Request
	server := fakeBarsServer(t, &reqs)
	defer server.Close()

	svc := NewTestMrktdataHandler(barsClient(server.URL), nil)

	resp, err := svc.GetHistoricalStockData(context.Background(), &mrktpb.HistoricalStockDataRequest{
		Symbols:   []string{"AAPL"},
		Timeframe: mrktpb.Timeframe_DAY,
	})
	require.NoError(t, err)
	require.Len(t, resp.SymbolBars, 1)
	assert.Equal(t, "AAPL", resp.SymbolBars[0].Symbol)
	require.Len(t, resp.SymbolBars[0].Bars, 2)

	bar := resp.SymbolBars[0].Bars[0]
	assert.Equal(t, float32(100.5), bar.Open)
	assert.Equal(t, float32(101.0), bar.High)
	assert.Equal(t, float32(99.5), bar.Low)
	assert.Equal(t, float32(100.75), bar.Close)
	assert.NotEmpty(t, bar.Time)

	require.NotEmpty(t, reqs)
	assert.Equal(t, "1Day", reqs[0].URL.Query().Get("timeframe"))
	assert.Equal(t, "sip", reqs[0].URL.Query().Get("feed"))
}

func TestGetHistoricalStockDataTimeframeMapping(t *testing.T) {
	cases := []struct {
		tf       mrktpb.Timeframe
		expected string
	}{
		{mrktpb.Timeframe_DAY, "1Day"},
		{mrktpb.Timeframe_HOUR, "1Hour"},
		{mrktpb.Timeframe_MIN, "1Min"},
	}
	for _, tc := range cases {
		t.Run(tc.expected, func(t *testing.T) {
			var reqs []*http.Request
			server := fakeBarsServer(t, &reqs)
			defer server.Close()

			svc := NewTestMrktdataHandler(barsClient(server.URL), nil)
			_, err := svc.GetHistoricalStockData(context.Background(), &mrktpb.HistoricalStockDataRequest{
				Symbols:   []string{"TSLA"},
				Timeframe: tc.tf,
			})
			require.NoError(t, err)
			require.NotEmpty(t, reqs)
			assert.Equal(t, tc.expected, reqs[0].URL.Query().Get("timeframe"))
		})
	}
}

func TestGetHistoricalStockDataUsesRequestedWindow(t *testing.T) {
	var reqs []*http.Request
	server := fakeBarsServer(t, &reqs)
	defer server.Close()

	svc := NewTestMrktdataHandler(barsClient(server.URL), nil)
	_, err := svc.GetHistoricalStockData(context.Background(), &mrktpb.HistoricalStockDataRequest{
		Symbols:   []string{"AAPL"},
		Timeframe: mrktpb.Timeframe_DAY,
		Start:     &mrktpb.Date{Year: 2026, Month: 1, Day: 2, Hour: 9, Min: 30},
		End:       &mrktpb.Date{Year: 2026, Month: 2, Day: 2, Hour: 16, Min: 0},
	})
	require.NoError(t, err)
	require.NotEmpty(t, reqs)

	start, err := time.Parse(time.RFC3339, reqs[0].URL.Query().Get("start"))
	require.NoError(t, err)
	assert.Equal(t, 2026, start.Year())
	assert.Equal(t, time.January, start.Month())
	assert.Equal(t, 2, start.Day())

	end, err := time.Parse(time.RFC3339, reqs[0].URL.Query().Get("end"))
	require.NoError(t, err)
	assert.Equal(t, time.February, end.Month())
}

func TestGetHistoricalStockDataMultipleSymbols(t *testing.T) {
	server := fakeBarsServer(t, nil)
	defer server.Close()

	svc := NewTestMrktdataHandler(barsClient(server.URL), nil)
	resp, err := svc.GetHistoricalStockData(context.Background(), &mrktpb.HistoricalStockDataRequest{
		Symbols:   []string{"AAPL", "TSLA", "NVDA"},
		Timeframe: mrktpb.Timeframe_HOUR,
	})
	require.NoError(t, err)
	require.Len(t, resp.SymbolBars, 3)
	assert.Equal(t, "AAPL", resp.SymbolBars[0].Symbol)
	assert.Equal(t, "TSLA", resp.SymbolBars[1].Symbol)
	assert.Equal(t, "NVDA", resp.SymbolBars[2].Symbol)
}

func TestGetHistoricalStockDataPropagatesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"rate limited"}`, http.StatusTooManyRequests)
	}))
	defer server.Close()

	svc := NewTestMrktdataHandler(barsClient(server.URL), nil)
	_, err := svc.GetHistoricalStockData(context.Background(), &mrktpb.HistoricalStockDataRequest{
		Symbols:   []string{"AAPL"},
		Timeframe: mrktpb.Timeframe_DAY,
	})
	assert.Error(t, err)
}

func TestGetHistoricalStockDataCachesWorkingFeed(t *testing.T) {
	var reqs []*http.Request
	server := fakeBarsServer(t, &reqs)
	defer server.Close()

	svc := NewTestMrktdataHandler(barsClient(server.URL), nil)

	for range 3 {
		_, err := svc.GetHistoricalStockData(context.Background(), &mrktpb.HistoricalStockDataRequest{
			Symbols:   []string{"AAPL"},
			Timeframe: mrktpb.Timeframe_DAY,
		})
		require.NoError(t, err)
	}

	require.Len(t, reqs, 3)
	for _, r := range reqs {
		assert.Equal(t, "sip", r.URL.Query().Get("feed"))
	}
}

func TestGetAvailableSymbolsFiltersAndCaches(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		assert.Contains(t, r.URL.Path, "/v2/assets")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"id":"1","symbol":"AAPL","name":"Apple Inc","tradable":true},
			{"id":"2","symbol":"DEAD","name":"Delisted Corp","tradable":false},
			{"id":"3","symbol":"TSLA","name":"Tesla Inc","tradable":true}
		]`))
	}))
	defer server.Close()

	svc := NewTestMrktdataHandler(nil, assetsClient(server.URL))

	resp, err := svc.GetAvailableSymbols(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)
	require.Len(t, resp.Symbols, 2, "non-tradable assets must be filtered out")
	assert.Equal(t, "AAPL", resp.Symbols[0].Name)
	assert.Equal(t, "Apple Inc", resp.Symbols[0].CompanyName)
	assert.Equal(t, "TSLA", resp.Symbols[1].Name)

	resp2, err := svc.GetAvailableSymbols(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)
	assert.Len(t, resp2.Symbols, 2)
	assert.Equal(t, int32(1), calls.Load(), "assets endpoint should be hit exactly once")
}

func TestGetAvailableSymbolsPropagatesError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"boom"}`, http.StatusInternalServerError)
	}))
	defer server.Close()

	svc := NewTestMrktdataHandler(nil, assetsClient(server.URL))
	_, err := svc.GetAvailableSymbols(context.Background(), &emptypb.Empty{})
	assert.Error(t, err)
}

func TestGetLatestPricesConvertsToCents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/trades/latest")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"trades":{
			"AAPL":{"t":"2026-06-11T15:00:00Z","p":189.505,"s":100},
			"TSLA":{"t":"2026-06-11T15:00:00Z","p":248.9,"s":50}
		}}`))
	}))
	defer server.Close()

	svc := NewTestMrktdataHandler(barsClient(server.URL), nil)
	resp, err := svc.GetLatestPrices(context.Background(), &mrktpb.LatestPricesRequest{
		Symbols: []string{"AAPL", "TSLA"},
	})
	require.NoError(t, err)
	require.Len(t, resp.Prices, 2)

	prices := map[string]int64{}
	for _, p := range resp.Prices {
		prices[p.Symbol] = p.PriceCents
	}
	assert.Equal(t, int64(18_951), prices["AAPL"], "fractional cents round half away from zero")
	assert.Equal(t, int64(24_890), prices["TSLA"])
}

func TestGetLatestPricesEmptyRequest(t *testing.T) {
	svc := NewTestMrktdataHandler(nil, nil)
	resp, err := svc.GetLatestPrices(context.Background(), &mrktpb.LatestPricesRequest{})
	require.NoError(t, err)
	assert.Empty(t, resp.Prices)
}

func TestGetLatestPricesPropagatesError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"boom"}`, http.StatusInternalServerError)
	}))
	defer server.Close()

	svc := NewTestMrktdataHandler(barsClient(server.URL), nil)
	_, err := svc.GetLatestPrices(context.Background(), &mrktpb.LatestPricesRequest{Symbols: []string{"AAPL"}})
	assert.Error(t, err)
}

func feedAwareBarsServer(t *testing.T, sipOK *atomic.Bool, capture *[]*http.Request) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if capture != nil {
			*capture = append(*capture, r.Clone(context.Background()))
		}
		if r.URL.Query().Get("feed") == "sip" && !sipOK.Load() {
			http.Error(w, `{"message":"subscription does not permit querying recent SIP data"}`, http.StatusForbidden)
			return
		}
		bars := map[string]any{}
		for _, symbol := range strings.Split(r.URL.Query().Get("symbols"), ",") {
			if symbol == "" {
				continue
			}
			bars[symbol] = []map[string]any{
				{"t": "2026-01-02T15:00:00Z", "o": 100.5, "h": 101.0, "l": 99.5, "c": 100.75, "v": 1000, "n": 10, "vw": 100.6},
			}
		}
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"next_page_token": nil, "bars": bars}))
	}))
}

func feedsOf(reqs []*http.Request) []string {
	out := make([]string, 0, len(reqs))
	for _, r := range reqs {
		out = append(out, r.URL.Query().Get("feed"))
	}
	return out
}

func fetchDay(t *testing.T, svc *MrktdataHandler) {
	t.Helper()
	_, err := svc.GetHistoricalStockData(context.Background(), &mrktpb.HistoricalStockDataRequest{
		Symbols:   []string{"AAPL"},
		Timeframe: mrktpb.Timeframe_DAY,
	})
	require.NoError(t, err)
}

func TestHistoricalFeedFallbackIsNonSticky(t *testing.T) {
	var sipOK atomic.Bool
	var reqs []*http.Request
	server := feedAwareBarsServer(t, &sipOK, &reqs)
	defer server.Close()

	svc := NewTestMrktdataHandler(barsClient(server.URL), nil)

	fetchDay(t, svc)
	assert.Equal(t, []string{"sip", "iex"}, feedsOf(reqs))
	assert.Equal(t, "iex", svc.histFeed)

	reqs = nil
	fetchDay(t, svc)
	assert.Equal(t, []string{"iex"}, feedsOf(reqs))

	sipOK.Store(true)
	svc.feedMu.Lock()
	svc.lastSipProbe = time.Now().Add(-sipProbeInterval - time.Minute)
	svc.feedMu.Unlock()

	reqs = nil
	fetchDay(t, svc)
	assert.Equal(t, []string{"sip"}, feedsOf(reqs))
	assert.Equal(t, "sip", svc.histFeed)

	reqs = nil
	fetchDay(t, svc)
	assert.Equal(t, []string{"sip"}, feedsOf(reqs))
}
