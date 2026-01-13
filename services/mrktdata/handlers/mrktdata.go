package handlers

import (
	"context"
	"log"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	mrktpb "github.com/yash-gadgil/glyph/services/gen/golang/mrktdata"
)

const historicalDelay = 16 * time.Minute

func (h *MrktdataHandler) GetHistoricalStockData(ctx context.Context, req *mrktpb.HistoricalStockDataRequest) (*mrktpb.HistoricalStockDataResponse, error) {
	var tf marketdata.TimeFrame
	var lookback time.Duration

	switch req.Timeframe {
	case mrktpb.Timeframe_HOUR:
		tf = marketdata.OneHour
		lookback = 30 * 24 * time.Hour
	case mrktpb.Timeframe_MIN:
		tf = marketdata.OneMin
		lookback = 7 * 24 * time.Hour
	default:
		tf = marketdata.OneDay
		lookback = 365 * 24 * time.Hour
	}

	end := time.Now().Add(-historicalDelay)
	start := end.Add(-lookback)

	if req.Start != nil {
		start = time.Date(int(req.Start.Year), time.Month(req.Start.Month), int(req.Start.Day), int(req.Start.Hour), int(req.Start.Min), 0, 0, time.UTC)
	}
	if req.End != nil {
		reqEnd := time.Date(int(req.End.Year), time.Month(req.End.Month), int(req.End.Day), int(req.End.Hour), int(req.End.Min), 0, 0, time.UTC)
		if reqEnd.Before(end) {
			end = reqEnd
		}
	}

	barsBySymbol, err := h.getMultiBarsWithFeed(req.Symbols, marketdata.GetBarsRequest{
		TimeFrame:  tf,
		Start:      start,
		End:        end,
		Adjustment: marketdata.Split,
	})
	if err != nil {
		log.Printf("error fetching bars for %v: %v", req.Symbols, err)
		return nil, err
	}

	response := &mrktpb.HistoricalStockDataResponse{}
	for _, symbol := range req.Symbols {
		sb := &mrktpb.SymbolBars{Symbol: symbol}
		for _, b := range barsBySymbol[symbol] {
			sb.Bars = append(sb.Bars, &mrktpb.Bar{
				Symbol: symbol,
				Open:   float32(b.Open),
				High:   float32(b.High),
				Low:    float32(b.Low),
				Close:  float32(b.Close),
				Volume: int64(b.Volume),
				Vwap:   float32(b.VWAP),
				Time:   b.Timestamp.UTC().Format(time.RFC3339),
			})
		}
		response.SymbolBars = append(response.SymbolBars, sb)
	}

	return response, nil
}

func (h *MrktdataHandler) getMultiBarsWithFeed(symbols []string, req marketdata.GetBarsRequest) (map[string][]marketdata.Bar, error) {
	feed, probe := h.feedDecision()
	if !probe {
		req.Feed = feed
		return h.stocksApi.GetMultiBars(symbols, req)
	}

	req.Feed = "sip"
	bars, err := h.stocksApi.GetMultiBars(symbols, req)
	if err == nil {
		h.adoptFeed("sip")
		return bars, nil
	}
	log.Printf("sip feed unavailable (%v); falling back to iex", err)

	req.Feed = "iex"
	bars, err = h.stocksApi.GetMultiBars(symbols, req)
	if err == nil {
		h.adoptFeed("iex")
	}
	return bars, err
}

func (h *MrktdataHandler) feedDecision() (feed string, probe bool) {
	h.feedMu.Lock()
	defer h.feedMu.Unlock()
	switch h.histFeed {
	case "sip":
		return "sip", false
	case "iex":
		if time.Since(h.lastSipProbe) < sipProbeInterval {
			return "iex", false
		}
		h.lastSipProbe = time.Now()
		return "iex", true
	default:
		h.lastSipProbe = time.Now()
		return "", true
	}
}

func (h *MrktdataHandler) adoptFeed(feed string) {
	h.feedMu.Lock()
	defer h.feedMu.Unlock()
	if h.histFeed != feed {
		log.Printf("historical feed: now serving %q (was %q)", feed, h.histFeed)
	}
	h.histFeed = feed
}
