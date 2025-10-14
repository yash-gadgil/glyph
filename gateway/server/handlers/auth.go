package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yash-gadgil/glyph/gateway/server/utils"
	authpb "github.com/yash-gadgil/glyph/services/gen/golang/auth"
)

func (cfg *Config) Register(w http.ResponseWriter, r *http.Request) {
	var req utils.AuthReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnErrorJSON(w, "Unable to parse Request", http.StatusBadRequest)
		return
	}

	serverAddr := cfg.AuthServiceAddr
	conn := utils.GetGrpcClient(serverAddr)
	defer conn.Close()

	c := authpb.NewAuthServiceClient(conn)
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*2)
	defer cancel()

	res, err := c.Register(ctx, &authpb.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		fmt.Println("Error in Server Response:", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (cfg *Config) Login(w http.ResponseWriter, r *http.Request) {
	var req utils.AuthReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnErrorJSON(w, "Unable to parse Request", http.StatusBadRequest)
		return
	}

	res := utils.AuthResp{
		RefreshToken: "wells",
		AccessToken:  "fargo",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (cfg *Config) OAuth(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	state := r.URL.Query().Get("state")

	serverAddr := cfg.AuthServiceAddr
	conn := utils.GetGrpcClient(serverAddr)
	defer conn.Close()

	c := authpb.NewAuthServiceClient(conn)
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*2)
	defer cancel()

	res, err := c.OAuthURL(ctx, &authpb.OAuthURLRequest{
		Provider: provider,
		Status:   state,
	})
	if err != nil {
		fmt.Println("Error in Server Response:", err)
	}

	fmt.Println(res.Url)
	http.Redirect(w, r, res.Url, http.StatusTemporaryRedirect)
}

func (cfg *Config) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	serverAddr := cfg.AuthServiceAddr
	conn := utils.GetGrpcClient(serverAddr)
	defer conn.Close()

	c := authpb.NewAuthServiceClient(conn)
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*2)
	defer cancel()

	res, err := c.OAuthCallback(ctx, &authpb.OAuthCallbackRequest{
		Code: code,
	})
	if err != nil {
		fmt.Println("Error in Server Response:", err)
	}
	fmt.Println(provider, state)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (cfg *Config) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	fmt.Println(token)
}
