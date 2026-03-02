package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/pkg/telemetry"
	"github.com/yash-gadgil/glyph/services/gateway/server/types"
	"github.com/yash-gadgil/glyph/services/gateway/server/utils"
	authpb "github.com/yash-gadgil/glyph/services/gen/golang/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (cfg *Config) LoadAuthRoutes(r chi.Router) {
	r.Post("/signup", cfg.Signup)
	r.Post("/signin", cfg.Signin)
	r.Post("/signout", cfg.Signout)
	r.Post("/forgot-password", cfg.ForgotPassword)
	r.Post("/reset-password", cfg.ResetPassword)
	r.Get("/oauth/{provider}", cfg.OAuth)
	r.Get("/oauth/{provider}/callback", cfg.OAuthCallback)
	r.Get("/verify", cfg.VerifyEmail)
	r.Get("/refresh", cfg.Refresh)
}

func (cfg *Config) Signup(w http.ResponseWriter, r *http.Request) {
	var req types.SignupReq

	log := logger.WithContextFields(r.Context(), cfg.log).With(
		logger.Action("signup"),
	)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("req_parse_error", logger.Stage("parse_request"), zap.Error(err))
		utils.ReturnErrorJSON(w, "unable to parse request", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	_, err := cfg.AuthClient.Signup(ctx, &authpb.SignupRequest{
		UserName: req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			log.Error("server_signup_error", logger.Stage("server_signup"), logger.KV("message", st.Message()), zap.Error(err))
			utils.ReturnErrorJSON(w, st.Message(), http.StatusInternalServerError)
			return
		}

		log.Error("server_signup_error", logger.Stage("server_signup"), zap.Error(err))
		utils.ReturnErrorJSON(w, "error signing up", http.StatusInternalServerError)
		return
	}

	log.Info("signup_email_sent")

	telemetry.SignupsTotal.Inc()
	w.WriteHeader(http.StatusOK)
}

func (cfg *Config) Signin(w http.ResponseWriter, r *http.Request) {
	var req types.SigninReq

	log := logger.WithContextFields(r.Context(), cfg.log).With(
		logger.Action("signin"),
	)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("req_parse_error", logger.Stage("parse_request"), zap.Error(err))
		utils.ReturnErrorJSON(w, "unable to parse request", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*2)
	defer cancel()

	res, err := cfg.AuthClient.Signin(ctx, &authpb.SigninRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		log.Error("server_signin_error", logger.Stage("server_signin"), zap.Error(err))

		st := status.Convert(err)
		switch st.Code() {
		case codes.NotFound:
			utils.ReturnErrorJSON(w, st.Message(), http.StatusNotFound)
		case codes.PermissionDenied:
			utils.ReturnErrorJSON(w, st.Message(), http.StatusForbidden)
		case codes.InvalidArgument:
			utils.ReturnErrorJSON(w, st.Message(), http.StatusBadRequest)
		default:
			utils.ReturnErrorJSON(w, "error signing in", http.StatusInternalServerError)
		}
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    res.RefreshToken,
		Expires:  time.Now().Add(time.Hour * 24 * 30),
		MaxAge:   int((time.Hour * 24 * 30).Seconds()),
		HttpOnly: true,
		Path:     "/",
		Secure:   SecureCookies(),
		SameSite: cookieSameSite(),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "accessToken",
		Value:    res.AccessToken,
		Expires:  time.Now().Add(time.Hour),
		MaxAge:   int(time.Hour.Seconds()),
		HttpOnly: true,
		Path:     "/",
		Secure:   SecureCookies(),
		SameSite: cookieSameSite(),
	})

	log.Info("user_signed_in")

	telemetry.SigninsTotal.Inc()
	w.WriteHeader(http.StatusOK)
}

func (cfg *Config) OAuth(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	state := r.URL.Query().Get("state")

	log := logger.WithContextFields(r.Context(), cfg.log).With(
		logger.Action("oauth"),
	)

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*2)
	defer cancel()

	res, err := cfg.AuthClient.OAuthURL(ctx, &authpb.OAuthURLRequest{
		Provider: provider,
		State:    state,
	})
	if err != nil {
		log.Error("url_fetch_error", logger.Stage("url_fetch"), zap.Error(err))
		utils.ReturnErrorJSON(w, "failed to fetch OAuth URL", http.StatusInternalServerError)
		return
	}

	log.Info("url_fetched")

	http.Redirect(w, r, res.Url, http.StatusTemporaryRedirect)
}

func (cfg *Config) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	log := logger.WithContextFields(r.Context(), cfg.log).With(
		logger.Action("oauth_callback"),
	)

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	res, err := cfg.AuthClient.OAuthCallback(ctx, &authpb.OAuthCallbackRequest{
		Code:     code,
		State:    state,
		Provider: provider,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			log.Error("server_callback_error", logger.Stage("callback_service"), logger.KV("message", st.Message()), zap.Error(err))
			utils.ReturnErrorJSON(w, st.Message(), http.StatusInternalServerError)
			return
		}

		log.Error("server_callback_error", logger.Stage("callback_service"), zap.Error(err))
		utils.ReturnErrorJSON(w, "OAuth callback failed", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    res.RefreshToken,
		Expires:  time.Now().Add(time.Hour * 24 * 7),
		HttpOnly: true,
		Path:     "/",
		Secure:   SecureCookies(),
		SameSite: cookieSameSite(),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "accessToken",
		Value:    res.AccessToken,
		Expires:  time.Now().Add(time.Minute * 30),
		HttpOnly: true,
		Path:     "/",
		Secure:   SecureCookies(),
		SameSite: cookieSameSite(),
	})

	frontend := os.Getenv("FRONTEND_URL")
	if frontend == "" {
		frontend = "http://localhost:3000"
	}
	frontend = strings.TrimRight(frontend, "/")

	log.Info("user_signed_in_redirecting")

	redirectTarget := fmt.Sprintf("%s/dashboard", frontend)
	http.Redirect(w, r, redirectTarget, http.StatusSeeOther)
}

func (cfg *Config) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	log := logger.WithContextFields(r.Context(), cfg.log).With(
		logger.Action("verify_email"),
	)
	frontend := os.Getenv("FRONTEND_URL")
	if frontend == "" {
		frontend = "http://localhost:3000"
	}

	if token == "" {
		log.Error("token_missing", logger.Stage("token_verification"), zap.Error(errors.New("missing token")))
		http.Redirect(w, r, strings.TrimRight(frontend, "/")+"/signin?error=invalid_link", http.StatusSeeOther)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*2)
	defer cancel()

	res, err := cfg.AuthClient.VerifyEmail(ctx, &authpb.VerificationRequest{
		Token: token,
	})
	if err != nil {
		log.Error("token_verification_error", logger.Stage("token_verification"), zap.Error(err))
		http.Redirect(w, r, strings.TrimRight(frontend, "/")+"/signin?error=verification_failed", http.StatusSeeOther)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    res.RefreshToken,
		Expires:  time.Now().Add(time.Hour * 24 * 7),
		HttpOnly: true,
		Path:     "/",
		Secure:   SecureCookies(),
		SameSite: cookieSameSite(),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "accessToken",
		Value:    res.AccessToken,
		Expires:  time.Now().Add(time.Minute * 30),
		HttpOnly: true,
		Path:     "/",
		Secure:   SecureCookies(),
		SameSite: cookieSameSite(),
	})

	log.Info("email_verified")

	http.Redirect(w, r, strings.TrimRight(frontend, "/")+"/dashboard", http.StatusSeeOther)
}

func (cfg *Config) Refresh(w http.ResponseWriter, r *http.Request) {

	log := logger.WithContextFields(r.Context(), cfg.log).With(
		logger.Action("refresh"),
	)

	cookie, err := r.Cookie("refreshToken")
	if err != nil {
		log.Error("cookie_not_found", logger.Stage("token_fetch"), zap.Error(err))
		utils.ReturnErrorJSON(w, "No refresh token found", http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	res, err := cfg.AuthClient.RefreshToken(ctx, &authpb.RefreshTokenRequest{
		RefreshToken: cookie.Value,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			log.Error("token_refresh_failed", logger.Stage("service_call"), logger.KV("message", st.Message()), zap.Error(err))
			utils.ReturnErrorJSON(w, "Invalid refresh token", http.StatusUnauthorized)
			return
		}

		log.Error("token_refresh_failed", logger.Stage("service_call"), zap.Error(err))
		utils.ReturnErrorJSON(w, "Refresh failed", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    res.RefreshToken,
		Expires:  time.Now().Add(time.Hour * 24 * 7),
		HttpOnly: true,
		Path:     "/",
		Secure:   SecureCookies(),
		SameSite: cookieSameSite(),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "accessToken",
		Value:    res.AccessToken,
		Expires:  time.Now().Add(time.Hour * 2),
		HttpOnly: true,
		Path:     "/",
		Secure:   SecureCookies(),
		SameSite: cookieSameSite(),
	})

	log.Info("token_refreshed")

	w.WriteHeader(http.StatusOK)
}

func (cfg *Config) Signout(w http.ResponseWriter, r *http.Request) {

	log := logger.WithContextFields(r.Context(), cfg.log).With(
		logger.Action("signout"),
	)

	http.SetCookie(w, &http.Cookie{
		Name:     "accessToken",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
		Secure:   SecureCookies(),
		SameSite: cookieSameSite(),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
		Secure:   SecureCookies(),
		SameSite: cookieSameSite(),
	})

	log.Info("user_signed_out")

	w.WriteHeader(http.StatusOK)
}

func (cfg *Config) ForgotPassword(w http.ResponseWriter, r *http.Request) {

	var req types.RecoveryReq

	log := logger.WithContextFields(r.Context(), cfg.log).With(
		logger.Action("forgot-password"),
	)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("req_parse_error", logger.Stage("parse_request"), zap.Error(err))
		utils.ReturnErrorJSON(w, "unable to parse request", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	_, err := cfg.AuthClient.ForgotPassword(ctx, &authpb.ForgotPasswordRequest{
		Email: req.Email,
	})
	if err != nil {
		log.Error("forgot_password_error", logger.Stage("service_call"), zap.Error(err))
	}

	log.Info("password_reset_requested")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "If an account exists with that email, a password reset link has been sent.",
	})
}

func (cfg *Config) ResetPassword(w http.ResponseWriter, r *http.Request) {

	var req types.ResetPasswordReq

	log := logger.WithContextFields(r.Context(), cfg.log).With(
		logger.Action("reset-password"),
	)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("req_parse_error", logger.Stage("parse_request"), zap.Error(err))
		utils.ReturnErrorJSON(w, "unable to parse request", http.StatusBadRequest)
		return
	}

	if req.Token == "" || req.NewPassword == "" {
		utils.ReturnErrorJSON(w, "Token and new password are required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	res, err := cfg.AuthClient.ResetPassword(ctx, &authpb.ResetPasswordRequest{
		Token:       req.Token,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			log.Error("reset_password_error", logger.Stage("service_call"), logger.KV("message", st.Message()), zap.Error(err))
			utils.ReturnErrorJSON(w, st.Message(), http.StatusBadRequest)
			return
		}

		log.Error("reset_password_error", logger.Stage("service_call"), zap.Error(err))
		utils.ReturnErrorJSON(w, "Failed to reset password", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    res.RefreshToken,
		Expires:  time.Now().Add(time.Hour * 24 * 30),
		MaxAge:   int((time.Hour * 24 * 30).Seconds()),
		HttpOnly: true,
		Path:     "/",
		Secure:   SecureCookies(),
		SameSite: cookieSameSite(),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "accessToken",
		Value:    res.AccessToken,
		Expires:  time.Now().Add(time.Hour),
		MaxAge:   int(time.Hour.Seconds()),
		HttpOnly: true,
		Path:     "/",
		Secure:   SecureCookies(),
		SameSite: cookieSameSite(),
	})

	log.Info("password_reset_complete")

	w.WriteHeader(http.StatusOK)
}
