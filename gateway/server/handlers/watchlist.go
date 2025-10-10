package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yash-gadgil/glyph/gateway/server/utils"
)

func GetWatchlists(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("your watchlists"))
}

func ConnectToWatchlist(w http.ResponseWriter, r *http.Request) {
	watchlistID := chi.URLParam(r, "id")
	symbol := r.URL.Query().Get("symbol")

	if symbol == "" {
		w.Write([]byte("watchlist no. " + watchlistID))
	} else {
		w.Write([]byte("symbol " + symbol + " from watchlist no. " + watchlistID))
	}
}

func CreateWatchlist(w http.ResponseWriter, r *http.Request) {
	var nw utils.Watchlist

	if err := json.NewDecoder(r.Body).Decode(&nw); err != nil {

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid watchlist parameter",
		})
		return
	}

	fmt.Println(nw.Name)

	w.Write([]byte("New watchlist made"))
}

func ModifyWatchlist(w http.ResponseWriter, r *http.Request) {
	watchlistID := chi.URLParam(r, "id")

	w.Write([]byte("edited watchlist no. " + watchlistID))
}

func DeleteWatchlist(w http.ResponseWriter, r *http.Request) {
	watchlistID := chi.URLParam(r, "id")

	w.Write([]byte("delete watchlist no. " + watchlistID))
}

func DeleteSymbolFromWatchlist(w http.ResponseWriter, r *http.Request) {
	watchlistID := chi.URLParam(r, "id")
	symbol := r.URL.Query().Get("symbol")

	w.Write([]byte("delete " + symbol + " from watchlist no. " + watchlistID))
}
