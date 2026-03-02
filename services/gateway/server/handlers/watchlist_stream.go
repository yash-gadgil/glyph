package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/pkg/telemetry"
	authpb "github.com/yash-gadgil/glyph/services/gen/golang/auth"
	mrktdata "github.com/yash-gadgil/glyph/services/gen/golang/mrktdata"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "" || origin == AllowedOrigin()
	},
}

func (cfg *Config) ConnectToWatchlist(w http.ResponseWriter, r *http.Request) {
	watchlistID := chi.URLParam(r, "id")

	userID, err := cfg.userIDFromCookie(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	watchlistCtx, watchlistCancel := context.WithTimeout(context.Background(), time.Second*2)
	defer watchlistCancel()

	watchlist, err := cfg.watchlistClient.GetWatchlist(watchlistCtx, &userpb.WatchlistSpecifier{
		Id:     watchlistID,
		UserId: userID,
	})
	if err != nil {
		http.Error(w, "Watchlist not found", http.StatusNotFound)
		return
	}

	cfg.streamSymbolsOverWS(w, r, watchlist.Symbols)
}

func (cfg *Config) ConnectToSymbols(w http.ResponseWriter, r *http.Request) {
	if _, err := cfg.userIDFromCookie(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	raw := strings.Split(r.URL.Query().Get("symbols"), ",")
	symbols := make([]string, 0, len(raw))
	for _, s := range raw {
		s = strings.ToUpper(strings.TrimSpace(s))
		if s != "" {
			symbols = append(symbols, s)
		}
	}
	if len(symbols) == 0 || len(symbols) > 50 {
		http.Error(w, "symbols query param required (max 50)", http.StatusBadRequest)
		return
	}

	cfg.streamSymbolsOverWS(w, r, symbols)
}

func (cfg *Config) userIDFromCookie(r *http.Request) (string, error) {
	accessCookie, err := r.Cookie("accessToken")
	if err != nil {
		return "", err
	}
	if cfg.AuthClient == nil {
		return "", status.Error(codes.Unavailable, "auth service unavailable")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	authRes, err := cfg.AuthClient.VerifyToken(ctx, &authpb.VerificationRequest{
		Token: accessCookie.Value,
	})
	if err != nil {
		return "", err
	}
	return authRes.UserId, nil
}

func (cfg *Config) streamSymbolsOverWS(w http.ResponseWriter, r *http.Request, symbols []string) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(logger.Action("watchlist_stream"))

	if cfg.mrktdataClient == nil {
		http.Error(w, "Market data service unavailable", http.StatusServiceUnavailable)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("ws_upgrade_failed", zap.Error(err))
		return
	}
	defer conn.Close()

	telemetry.WSConnectionsActive.Inc()
	defer telemetry.WSConnectionsActive.Dec()

	streamCtx, streamCancel := context.WithCancel(context.Background())
	defer streamCancel()

	stream, err := cfg.mrktdataClient.WatchlistStream(streamCtx)
	if err != nil {
		log.Error("mrktdata_stream_start_failed", zap.Error(err))
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Stream error"))
		return
	}

	if err := stream.Send(&mrktdata.WatchlistStreamRequest{
		Action:  mrktdata.WatchlistStreamRequest_SUBSCRIBE,
		Symbols: symbols,
	}); err != nil {
		log.Error("mrktdata_subscribe_failed", zap.Error(err))
		return
	}

	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				streamCancel()
				return
			}
		}
	}()

	for {
		bar, err := stream.Recv()
		if err != nil {
			if streamCtx.Err() == nil {
				log.Error("mrktdata_stream_recv_failed", zap.Error(err))
				conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Stream error"))
			}
			return
		}
		if err := conn.WriteJSON(bar); err != nil {
			return
		}
	}
}
