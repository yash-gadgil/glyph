package handlers

import (
	"log"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

// SERVICE ADDRESS CONFIG ------------------------------------------------------------

type Config struct {
	GatewayServiceAddr  string
	AuthServiceAddr     string
	UserServiceAddr     string
	MrktdataServiceAddr string
}

func NewFromEnv() *Config {
	cfg := &Config{
		GatewayServiceAddr:  "",
		AuthServiceAddr:     "",
		UserServiceAddr:     "",
		MrktdataServiceAddr: "",
	}
	// Try common .env locations; don't exit fatally if not found.
	// Order: current dir, parent dir, filesystem root.
	// TECH DEBT
	if err := godotenv.Load(".env"); err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			if err := godotenv.Load("/.env"); err != nil {
				log.Println("Warning: .env not found at .env, ../.env, or /.env; relying on existing environment")
			}
		}
	}
	if v := os.Getenv("GATEWAY_SVC_PORT"); v != "" {
		cfg.GatewayServiceAddr = v
	}
	if v := os.Getenv("AUTH_SVC_PORT"); v != "" {
		cfg.AuthServiceAddr = v
	}
	if v := os.Getenv("USER_SVC_PORT"); v != "" {
		cfg.UserServiceAddr = v
	}
	if v := os.Getenv("MRKETDATA_SVC_PORT"); v != "" {
		cfg.MrktdataServiceAddr = v
	}
	return cfg
}

// ORDER ROUTES ----------------------------------------------------------------------

func (cfg *Config) LoadOrderRoutes(r chi.Router) {

	r.Get("/", cfg.GetOrders)

	r.Post("/", cfg.CreateOrder)

	r.Delete("/{id}", cfg.DeleteOrder)
}

// PORTFOLIO ROUTES ----------------------------------------------------------------------

func (cfg *Config) LoadPortfolioRoutes(r chi.Router) {

	r.Get("/", cfg.GetPortfolio)

	r.Get("/holdings", cfg.GetHoldings)

	r.Get("/positions", cfg.GetPositions)
}

// WATCHLIST ROUTES ----------------------------------------------------------------------

func (cfg *Config) LoadWatchlistRoutes(r chi.Router) {

	r.Get("/", cfg.GetWatchlists)

	r.Get("/{id}", cfg.ConnectToWatchlist)

	r.Post("/", cfg.CreateWatchlist)

	r.Patch("/{id}", cfg.ModifyWatchlist)

	r.Delete("/{id}", cfg.DeleteWatchlist)

	r.Delete("/{id}", cfg.DeleteSymbolFromWatchlist)

	/* r.Post("/{id}/symbol", func(w http.ResponseWriter, r *http.Request) {
		watchlistID := chi.URLParam(r, "id")

		w.Write([]byte("watchlist no. " + watchlistID))
	}) */
}

// ACCOUNT ROUTES ----------------------------------------------------------------------

func (cfg *Config) LoadAccountRoutes(r chi.Router) {

	r.Get("/", cfg.GetAccount)

	r.Get("/funds", cfg.GetFunds)

	r.Get("/profile", cfg.GetProfile)

	r.Get("/trades", cfg.GetTrades)
}

// AUTH ROUTES ---------------------------------------------------------------------------

func (cfg *Config) LoadAuthRoutes(r chi.Router) {

	r.Post("/register", cfg.Register)

	r.Post("/login", cfg.Login)

	r.Get("/oauth/{provider}", cfg.OAuth)

	r.Get("/oauth/{provider}/callback", cfg.OAuthCallback)

	r.Post("/verify", cfg.VerifyEmail)

}
