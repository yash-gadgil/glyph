package main

import (
	"log"

	"github.com/yash-gadgil/glyph/gateway/server"
)

func main() {

	cfg := server.LoadConfigFromYML("services.yml")
	srv := server.NewServer()

	srv.
		AddRoute("/orders", cfg.LoadOrderRoutes).
		AddRoute("/portfolio", cfg.LoadPortfolioRoutes).
		AddRoute("/watchlists", cfg.LoadWatchlistRoutes).
		AddRoute("/account", cfg.LoadAccountRoutes)

	if err := srv.ServeAtPort(":3000"); err != nil {
		log.Fatal("Error Starting Server")
	}

}
