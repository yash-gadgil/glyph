package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yash-gadgil/glyph/gateway/server/utils"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
)

func (cfg *Config) GetWatchlists(w http.ResponseWriter, r *http.Request) {

	serverAddr := cfg.UserServiceAddr
	conn := utils.GetGrpcClient(serverAddr)
	defer conn.Close()

	c := userpb.NewWatchlistServiceClient(conn)
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*2)
	defer cancel()

	res, err := c.GetWatchlists(ctx, &userpb.WatchlistsRequest{
		UserID: 253,
	})
	if err != nil {
		fmt.Println("Error in Server Response:", err)
	}

	//w.Write([]byte("your watchlists"))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (cfg *Config) ConnectToWatchlist(w http.ResponseWriter, r *http.Request) {
	watchlistID := chi.URLParam(r, "id")
	symbol := r.URL.Query().Get("symbol")

	if symbol == "" {
		w.Write([]byte("watchlist no. " + watchlistID))
	} else {
		w.Write([]byte("symbol " + symbol + " from watchlist no. " + watchlistID))
	}
}

func (cfg *Config) CreateWatchlist(w http.ResponseWriter, r *http.Request) {
	var nw utils.Watchlist

	if err := json.NewDecoder(r.Body).Decode(&nw); err != nil {
		utils.ReturnErrorJSON(w, "Invalid watchlist parameter", http.StatusBadRequest)
		return
	}

	fmt.Println(nw.Name)

	w.Write([]byte("New watchlist made"))
}

func (cfg *Config) ModifyWatchlist(w http.ResponseWriter, r *http.Request) {
	watchlistID := chi.URLParam(r, "id")

	w.Write([]byte("edited watchlist no. " + watchlistID))
}

func (cfg *Config) DeleteWatchlist(w http.ResponseWriter, r *http.Request) {
	watchlistID := chi.URLParam(r, "id")

	w.Write([]byte("delete watchlist no. " + watchlistID))
}

func (cfg *Config) DeleteSymbolFromWatchlist(w http.ResponseWriter, r *http.Request) {
	watchlistID := chi.URLParam(r, "id")
	symbol := r.URL.Query().Get("symbol")

	w.Write([]byte("delete " + symbol + " from watchlist no. " + watchlistID))
}
