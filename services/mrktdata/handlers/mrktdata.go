package handlers

import (
	"context"
	"io"
	"log"
	"math"
	"strconv"
	"time"

	alpaca "github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	mrktpb "github.com/yash-gadgil/glyph/services/gen/golang/mrktdata"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

const historicalDelay = 16 * time.Minute

const defaultNewsLimit = 12

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

func (h *MrktdataHandler) WatchlistStream(stream grpc.BidiStreamingServer[mrktpb.WatchlistStreamRequest, mrktpb.MarketUpdate]) error {
	sub := h.hub.Register()
	defer h.hub.Unregister(sub)

	sendErr := make(chan error, 1)
	go func() {
		for bar := range sub.ch {
			if err := stream.Send(&mrktpb.MarketUpdate{SymbolBar: []*mrktpb.Bar{bar}}); err != nil {
				sendErr <- err
				return
			}
		}
	}()

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			select {
			case serr := <-sendErr:
				return serr
			default:
			}
			return err
		}

		switch msg.Action {
		case mrktpb.WatchlistStreamRequest_UNSUBSCRIBE:
			if err := h.hub.SetSymbols(sub, nil); err != nil {
				log.Printf("watchlist stream unsubscribe failed: %v", err)
				return err
			}
		default:
			if err := h.hub.SetSymbols(sub, msg.Symbols); err != nil {
				log.Printf("watchlist stream subscribe failed: %v", err)
				return err
			}
		}
	}
}

func (h *MrktdataHandler) GetNews(ctx context.Context, req *mrktpb.NewsRequest) (*mrktpb.NewsResponse, error) {
	limit := int(req.Limit)
	if limit <= 0 {
		limit = defaultNewsLimit
	}

	articles, err := h.stocksApi.GetNews(marketdata.GetNewsRequest{
		Symbols:    req.Symbols,
		TotalLimit: limit,
	})
	if err != nil {
		return nil, err
	}

	resp := &mrktpb.NewsResponse{}
	for _, a := range articles {
		article := &mrktpb.NewsArticle{
			Id:        strconv.Itoa(a.ID),
			Headline:  a.Headline,
			Summary:   a.Summary,
			Source:    a.Author,
			Url:       a.URL,
			Symbols:   a.Symbols,
			CreatedAt: a.CreatedAt.Format(time.RFC3339),
		}
		if len(a.Images) > 0 {
			article.ImageUrl = a.Images[0].URL
		}
		resp.Articles = append(resp.Articles, article)
	}
	return resp, nil
}

func (h *MrktdataHandler) GetLatestPrices(ctx context.Context, req *mrktpb.LatestPricesRequest) (*mrktpb.LatestPricesResponse, error) {
	if len(req.Symbols) == 0 {
		return &mrktpb.LatestPricesResponse{}, nil
	}

	trades, err := h.stocksApi.GetLatestTrades(req.Symbols, marketdata.GetLatestTradeRequest{
		Feed: "iex",
	})
	if err != nil {
		return nil, err
	}

	resp := &mrktpb.LatestPricesResponse{}
	for symbol, trade := range trades {
		resp.Prices = append(resp.Prices, &mrktpb.SymbolPrice{
			Symbol:     symbol,
			PriceCents: int64(math.Round(trade.Price * 100)),
		})
	}
	return resp, nil
}

func (h *MrktdataHandler) GetAvailableSymbols(ctx context.Context, _ *emptypb.Empty) (*mrktpb.AvailableSymbolsResponse, error) {
	h.symbolsMu.RLock()
	cached := h.symbolsCache
	h.symbolsMu.RUnlock()

	if cached == nil {
		res, err := h.alpacaClient.GetAssets(alpaca.GetAssetsRequest{
			Status:     "active",
			AssetClass: "us_equity",
		})
		if err != nil {
			log.Println("error getting symbols", err)
			return nil, err
		}
		newCache := make([]cachedSymbol, 0, len(res))
		for _, asset := range res {
			if asset.Tradable {
				newCache = append(newCache, cachedSymbol{
					Symbol:      asset.Symbol,
					CompanyName: asset.Name,
				})
			}
		}
		h.symbolsMu.Lock()
		h.symbolsCache = newCache
		h.symbolsMu.Unlock()
		cached = newCache
	}

	resp := &mrktpb.AvailableSymbolsResponse{}
	for _, sym := range cached {
		resp.Symbols = append(resp.Symbols, &mrktpb.Symbol{
			Name:        sym.Symbol,
			CompanyName: sym.CompanyName,
		})
	}
	return resp, nil
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
