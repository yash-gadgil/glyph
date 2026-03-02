package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/services/gateway/server/types"
	"github.com/yash-gadgil/glyph/services/gateway/server/utils"
	mrktdata "github.com/yash-gadgil/glyph/services/gen/golang/mrktdata"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (cfg *Config) LoadWatchlistRoutes(r chi.Router) {
	r.Use(cfg.AuthMiddleware)

	r.Get("/", cfg.GetWatchlists)
	r.Get("/symbols", cfg.GetAvailableSymbols)
	r.Post("/history", cfg.GetHistoricalData)
	r.Get("/stream", cfg.ConnectToSymbols)
	r.Get("/{id}/info", cfg.GetWatchlistInfo)
	r.Get("/{id}", cfg.ConnectToWatchlist)
	r.Post("/", cfg.CreateWatchlist)
	r.Patch("/{id}", cfg.ModifyWatchlist)
	r.Delete("/{id}", cfg.DeleteWatchlist)
}

func (cfg *Config) GetWatchlists(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(
		logger.Action("get_watchlists"),
	)

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		log.Error("user_id_missing", logger.Stage("context_extraction"))
		utils.ReturnErrorJSON(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	log = log.With(logger.KV("user_id", userID))

	if cfg.watchlistClient == nil {
		utils.ReturnErrorJSON(w, "Watchlist service unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*2)
	defer cancel()

	res, err := cfg.watchlistClient.GetWatchlists(ctx, &userpb.UserSpecifier{
		UserId: userID,
	})
	if err != nil {
		log.Error("watchlist_fetch_error", logger.Stage("fetch_watchlists"), zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to retrieve watchlists", http.StatusInternalServerError)
		return
	}

	log.Info("fetched_watchlists", logger.Stage("success"))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (cfg *Config) GetWatchlistInfo(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(logger.Action("get_watchlist_info"))

	watchlistID := chi.URLParam(r, "id")
	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		utils.ReturnErrorJSON(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}
	if cfg.watchlistClient == nil {
		utils.ReturnErrorJSON(w, "Watchlist service unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*2)
	defer cancel()

	watchlist, err := cfg.watchlistClient.GetWatchlist(ctx, &userpb.WatchlistSpecifier{
		Id:     watchlistID,
		UserId: userID,
	})
	if err != nil {
		log.Error("watchlist_fetch_error", logger.KV("watchlist_id", watchlistID), zap.Error(err))
		utils.ReturnErrorJSON(w, "Watchlist not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(watchlist)
}

func (cfg *Config) GetAvailableSymbols(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if len(q) < 1 {
		utils.ReturnErrorJSON(w, "Query parameter 'q' is required (min 1 character)", http.StatusBadRequest)
		return
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	if cfg.mrktdataClient == nil {
		utils.ReturnErrorJSON(w, "Market data service unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	res, err := cfg.mrktdataClient.GetAvailableSymbols(ctx, &emptypb.Empty{})
	if err != nil {
		utils.ReturnErrorJSON(w, "Unable to get available symbols", http.StatusInternalServerError)
		return
	}

	upper := strings.ToUpper(q)
	lower := strings.ToLower(q)

	type scored struct {
		sym   *mrktdata.Symbol
		score int
	}
	var matches []scored

	for _, sym := range res.Symbols {
		name := sym.Name
		company := strings.ToLower(sym.CompanyName)

		var score int
		switch {
		case name == upper:
			score = 100
		case strings.HasPrefix(name, upper):
			score = 50
		case strings.HasPrefix(company, lower):
			score = 30
		case strings.Contains(company, lower):
			score = 10
		default:
			continue
		}
		matches = append(matches, scored{sym: sym, score: score})
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score > matches[j].score
		}
		return matches[i].sym.Name < matches[j].sym.Name
	})

	if len(matches) > limit {
		matches = matches[:limit]
	}

	filtered := make([]*mrktdata.Symbol, len(matches))
	for i, m := range matches {
		filtered[i] = m.sym
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&mrktdata.AvailableSymbolsResponse{Symbols: filtered})
}

func (cfg *Config) GetHistoricalData(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(logger.Action("get_historical_data"))

	var req struct {
		Symbols   []string `json:"symbols"`
		Timeframe string   `json:"timeframe"`
		Days      int      `json:"days"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnErrorJSON(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Symbols) == 0 {
		utils.ReturnErrorJSON(w, "At least one symbol is required", http.StatusBadRequest)
		return
	}

	var tf mrktdata.Timeframe
	switch req.Timeframe {
	case "HOUR":
		tf = mrktdata.Timeframe_HOUR
	case "MIN":
		tf = mrktdata.Timeframe_MIN
	default:
		tf = mrktdata.Timeframe_DAY
	}

	grpcReq := &mrktdata.HistoricalStockDataRequest{
		Symbols:   req.Symbols,
		Timeframe: tf,
	}
	if req.Days > 0 && req.Days <= 366*5 {
		start := time.Now().UTC().AddDate(0, 0, -req.Days)
		grpcReq.Start = &mrktdata.Date{
			Year:  int32(start.Year()),
			Month: int32(start.Month()),
			Day:   int32(start.Day()),
		}
	}

	if cfg.mrktdataClient == nil {
		utils.ReturnErrorJSON(w, "Market data service unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
	defer cancel()

	res, err := cfg.mrktdataClient.GetHistoricalStockData(ctx, grpcReq)
	if err != nil {
		log.Error("historical_fetch_error", zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to fetch historical data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (cfg *Config) CreateWatchlist(w http.ResponseWriter, r *http.Request) {
	var nw types.Watchlist

	log := logger.WithContextFields(r.Context(), cfg.log).With(
		logger.Action("create_watchlist"),
	)

	if err := json.NewDecoder(r.Body).Decode(&nw); err != nil {
		log.Error("invalid_request", logger.Stage("request_parse"), zap.Error(err))
		utils.ReturnErrorJSON(w, "Invalid watchlist parameter", http.StatusBadRequest)
		return
	}

	if cfg.watchlistClient == nil {
		utils.ReturnErrorJSON(w, "Watchlist service unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*2)
	defer cancel()

	res, err := cfg.watchlistClient.CreateWatchlist(ctx, &userpb.CreateWatchlistRequest{
		Name:   &nw.Name,
		UserId: r.Context().Value(userIDKey).(string),
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.AlreadyExists:
				log.Error("name_already_taken", logger.Stage("service_call"), logger.KV("message", st.Message()), zap.Error(err))
				utils.ReturnErrorJSON(w, st.Message(), http.StatusConflict)
				return
			case codes.InvalidArgument:
				log.Error("invalid_argument", logger.Stage("service_call"), logger.KV("message", st.Message()), zap.Error(err))
				utils.ReturnErrorJSON(w, st.Message(), http.StatusBadRequest)
				return
			}
		}

		log.Error("service_error", logger.Stage("service_call"), zap.Error(err))
		utils.ReturnErrorJSON(w, "Error creating watchlist", http.StatusInternalServerError)
		return
	}

	log.Info("created_watchlist", logger.KV("name", nw.Name))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (cfg *Config) firstUnknownSymbol(parent context.Context, symbols []string) (string, bool) {
	if len(symbols) == 0 || cfg.mrktdataClient == nil {
		return "", false
	}

	ctx, cancel := context.WithTimeout(parent, time.Second*3)
	defer cancel()

	res, err := cfg.mrktdataClient.GetAvailableSymbols(ctx, &emptypb.Empty{})
	if err != nil {
		return "", false
	}

	known := make(map[string]struct{}, len(res.Symbols))
	for _, sym := range res.Symbols {
		known[strings.ToUpper(sym.Name)] = struct{}{}
	}

	for _, s := range symbols {
		if _, ok := known[strings.ToUpper(strings.TrimSpace(s))]; !ok {
			return s, true
		}
	}
	return "", false
}

func (cfg *Config) ModifyWatchlist(w http.ResponseWriter, r *http.Request) {
	watchlistID := chi.URLParam(r, "id")

	var action userpb.ModifyWatchlistRequest_Action
	switch r.URL.Query().Get("action") {
	case "subscribe":
		action = userpb.ModifyWatchlistRequest_SUBSCRIBE
	case "unsubscribe":
		action = userpb.ModifyWatchlistRequest_UNSUBSCRIBE
	default:
		utils.ReturnErrorJSON(w, "Invalid modification action", http.StatusBadRequest)
		return
	}

	var symbolArr types.SymbolsArr
	if err := json.NewDecoder(r.Body).Decode(&symbolArr); err != nil {
		utils.ReturnErrorJSON(w, "Unable to parse symbols", http.StatusBadRequest)
		return
	}

	if cfg.watchlistClient == nil {
		utils.ReturnErrorJSON(w, "Watchlist service unavailable", http.StatusServiceUnavailable)
		return
	}

	if action == userpb.ModifyWatchlistRequest_SUBSCRIBE {
		if bad, ok := cfg.firstUnknownSymbol(r.Context(), symbolArr.Symbols); ok {
			utils.ReturnErrorJSON(w, "Unknown or unsupported symbol: "+bad, http.StatusBadRequest)
			return
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*2)
	defer cancel()

	res, err := cfg.watchlistClient.ModifyWatchlist(ctx, &userpb.ModifyWatchlistRequest{
		Action:  action,
		UserId:  r.Context().Value(userIDKey).(string),
		Id:      watchlistID,
		Symbols: symbolArr.Symbols,
	})
	if err != nil {
		utils.ReturnErrorJSON(w, "Unable to modify Watchlist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (cfg *Config) DeleteWatchlist(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(logger.Action("delete_watchlist"))

	watchlistID := chi.URLParam(r, "id")

	if cfg.watchlistClient == nil {
		utils.ReturnErrorJSON(w, "Watchlist service unavailable", http.StatusServiceUnavailable)
		return
	}

	if symbol := r.URL.Query().Get("symbol"); symbol != "" {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*2)
		defer cancel()
		_, err := cfg.watchlistClient.DeleteSymbolFromWatchlist(ctx, &userpb.DeleteSymbolRequest{
			Id:     watchlistID,
			UserId: r.Context().Value(userIDKey).(string),
			Symbol: symbol,
		})
		if err != nil {
			log.Error("delete_symbol_error", zap.Error(err))
			utils.ReturnErrorJSON(w, "Error deleting symbol", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*2)
	defer cancel()

	_, err := cfg.watchlistClient.DeleteWatchlist(ctx, &userpb.WatchlistSpecifier{
		Id:     watchlistID,
		UserId: r.Context().Value(userIDKey).(string),
	})
	if err != nil {
		log.Error("delete_watchlist_error", zap.Error(err))
		utils.ReturnErrorJSON(w, "Error deleting watchlist", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
