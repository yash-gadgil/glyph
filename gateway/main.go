package main

import (
	"log"

	"github.com/yash-gadgil/glyph/gateway/server"
	"github.com/yash-gadgil/glyph/gateway/server/handlers"
)

func main() {

	cfg := handlers.NewFromEnv()
	srv := server.NewServer()

	srv.
		AddRoute("/orders", cfg.LoadOrderRoutes).
		AddRoute("/portfolio", cfg.LoadPortfolioRoutes).
		AddRoute("/watchlists", cfg.LoadWatchlistRoutes).
		AddRoute("/account", cfg.LoadAccountRoutes).
		AddRoute("/auth", cfg.LoadAuthRoutes)

	if err := srv.ServeAtPort(cfg.GatewayServiceAddr); err != nil {
		log.Fatal("Error Starting Server")
	}

}
