package server

import (
	"log"

	"github.com/go-chi/chi/v5"
	"github.com/spf13/viper"
	"github.com/yash-gadgil/glyph/gateway/server/handlers"
)

// SERVICE ADDRESS CONFIG ------------------------------------------------------------

type ServiceAddr struct {
	Host string
	Port string
}

type Config struct {
	Services map[string]ServiceAddr
}

func LoadConfigFromYML(path string) *Config {
	viper.SetConfigFile(path)
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	var cfg Config
	if err = viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Unable to decode config: %v", err)
	}

	return &cfg
}

// ORDER ROUTES ----------------------------------------------------------------------

func (cfg *Config) LoadOrderRoutes(r chi.Router) {

	r.Get("/", handlers.GetOrders)

	r.Post("/", handlers.CreateOrder)

	r.Delete("/{id}", handlers.DeleteOrder)
}

// PORTFOLIO ROUTES ----------------------------------------------------------------------

func (cfg *Config) LoadPortfolioRoutes(r chi.Router) {

	r.Get("/", handlers.GetPortfolio)

	r.Get("/holdings", handlers.GetHoldings)

	r.Get("/positions", handlers.GetPositions)
}

// WATCHLIST ROUTES ----------------------------------------------------------------------

func (cfg *Config) LoadWatchlistRoutes(r chi.Router) {

	r.Get("/", handlers.GetWatchlists)

	r.Get("/{id}", handlers.ConnectToWatchlist)

	r.Post("/", handlers.CreateWatchlist)

	r.Patch("/{id}", handlers.ModifyWatchlist)

	r.Delete("/{id}", handlers.DeleteWatchlist)

	r.Delete("/{id}", handlers.DeleteSymbolFromWatchlist)

	/* r.Post("/{id}/symbol", func(w http.ResponseWriter, r *http.Request) {
		watchlistID := chi.URLParam(r, "id")

		w.Write([]byte("watchlist no. " + watchlistID))
	}) */
}

// ACCOUNT ROUTES ----------------------------------------------------------------------

func (cfg *Config) LoadAccountRoutes(r chi.Router) {

	r.Get("/", handlers.GetAccount)

	r.Get("/funds", handlers.GetFunds)

	r.Get("/profile", handlers.GetProfile)

	r.Get("/trades", handlers.GetTrades)
}
