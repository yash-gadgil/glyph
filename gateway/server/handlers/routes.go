package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yash-gadgil/glyph/gateway/server/utils"
	authpb "github.com/yash-gadgil/glyph/services/gen/golang/auth"
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

	fmt.Println("Port:", os.Getenv("GATEWAY_SVC_PORT"))
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

// HELPER FUNCTIONS ---------------------------------------------------------------------------------

type contextKey string

const userIDKey contextKey = "userID"

func extractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("invalid authorization format, expected 'Bearer <token>'")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", errors.New("empty token")
	}

	return token, nil
}

// MIDDLEWARE ---------------------------------------------------------------------------------------

func (cfg *Config) isAuthenticated(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		token, err := extractBearerToken(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		serverAddr := cfg.AuthServiceAddr
		conn := utils.GetGrpcClient(serverAddr)
		defer conn.Close()

		c := authpb.NewAuthServiceClient(conn)
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*2)
		defer cancel()

		res, err := c.VerifyToken(ctx, &authpb.VerificationRequest{Token: token})
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx = context.WithValue(r.Context(), userIDKey, res.UserID)
		next(w, r.WithContext(ctx))
	}
}

// ORDER ROUTES ----------------------------------------------------------------------

func (cfg *Config) LoadOrderRoutes(r chi.Router) {

	r.Get("/", cfg.isAuthenticated(cfg.GetOrders))

	r.Post("/", cfg.isAuthenticated(cfg.CreateOrder))

	r.Delete("/{id}", cfg.isAuthenticated(cfg.DeleteOrder))
}

// PORTFOLIO ROUTES ----------------------------------------------------------------------

func (cfg *Config) LoadPortfolioRoutes(r chi.Router) {

	r.Get("/", cfg.isAuthenticated(cfg.GetPortfolio))

	r.Get("/holdings", cfg.isAuthenticated(cfg.GetHoldings))

	r.Get("/positions", cfg.isAuthenticated(cfg.GetPositions))
}

// WATCHLIST ROUTES ----------------------------------------------------------------------

func (cfg *Config) LoadWatchlistRoutes(r chi.Router) {

	r.Get("/", cfg.isAuthenticated(cfg.GetWatchlists))

	r.Get("/{id}", cfg.isAuthenticated(cfg.ConnectToWatchlist))

	r.Post("/", cfg.isAuthenticated(cfg.CreateWatchlist))

	r.Patch("/{id}", cfg.isAuthenticated(cfg.ModifyWatchlist))

	r.Delete("/{id}", cfg.isAuthenticated(cfg.DeleteWatchlist))

	r.Delete("/{id}", cfg.isAuthenticated(cfg.DeleteSymbolFromWatchlist))
}

// ACCOUNT ROUTES ----------------------------------------------------------------------

func (cfg *Config) LoadAccountRoutes(r chi.Router) {

	r.Get("/", cfg.isAuthenticated(cfg.GetAccount))

	r.Get("/funds", cfg.isAuthenticated(cfg.GetFunds))

	r.Get("/profile", cfg.isAuthenticated(cfg.GetProfile))

	r.Get("/trades", cfg.isAuthenticated(cfg.GetTrades))
}

// AUTH ROUTES ---------------------------------------------------------------------------

func (cfg *Config) LoadAuthRoutes(r chi.Router) {

	r.Post("/register", cfg.Register)

	r.Post("/login", cfg.Login)

	r.Get("/oauth/{provider}", cfg.OAuth)

	r.Get("/oauth/{provider}/callback", cfg.OAuthCallback)

	r.Get("/verify", cfg.VerifyEmail)

}
