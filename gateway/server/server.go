package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "google.golang.org/grpc"
)

type Server struct {
	mx *chi.Mux
}

func NewServer() *Server {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"service": "api-gateway",
			"version": "1.0.0",
		})
	})

	return &Server{
		mx: r,
	}
}

func (s *Server) ServeAtPort(port string) error {
	fmt.Println("Starting Server...")
	return http.ListenAndServe(port, s.mx)
}

func (s *Server) AddRoute(routeName string, handlerFunc func(chi.Router)) *Server {
	s.mx.Route(routeName, handlerFunc)
	return s
}
