package handlers

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"math/big"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/services/gateway/server/utils"
	"go.uber.org/zap"
)

type contextKey string

const userIDKey contextKey = "userID"
const deviceIDKey contextKey = "deviceID"
const deviceIDCookie = "device_id"

func AllowedOrigin() string {
	if v := os.Getenv("GATEWAY_ALLOWED_ORIGIN"); v != "" {
		return v
	}
	return "http://localhost:3000"
}

func SecureCookies() bool {
	return os.Getenv("GATEWAY_SECURE_COOKIES") == "true"
}

func cookieSameSite() http.SameSite {
	if SecureCookies() {
		return http.SameSiteNoneMode
	}
	return http.SameSiteLaxMode
}

func CORSMiddleware(next http.Handler) http.Handler {
	allowed := AllowedOrigin()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowed)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (cfg *Config) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		log := logger.WithContextFields(r.Context(), cfg.log).With(
			logger.Action("auth_middleware"),
		)

		cookie, err := r.Cookie("accessToken")
		if err != nil {
			log.Error("cookie_fetch_error", logger.Stage("cookie_fetch"), zap.Error(err))
			utils.ReturnErrorJSON(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, jwt.ErrSignatureInvalid
			}

			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, jwt.ErrTokenMalformed
			}

			return cfg.jwks.getKey(r.Context(), kid)
		})

		if err != nil {
			log.Error("token_parse_failed", logger.Stage("token_validation"), zap.Error(err))
			utils.ReturnErrorJSON(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			log.Error("token_invalid", logger.Stage("token_validation"), zap.Error(errors.New("invalid token")))
			utils.ReturnErrorJSON(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Error("invalid_token_claims", logger.Stage("claims_validation"), zap.Error(errors.New("invalid token claims")))
			utils.ReturnErrorJSON(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		userID, ok := claims["userid"].(string)
		if !ok {
			log.Error("user_id_invalid", logger.Stage("user_id_validation"), zap.Error(errors.New("invalid user id")))
			utils.ReturnErrorJSON(w, "Invalid user ID in token", http.StatusUnauthorized)
			return
		}

		log.Info("user_authenticated", logger.Stage("success"), logger.KV("user_id", userID))

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func jwksToRSAPublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, err
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}, nil
}

func (cfg *Config) DeviceIDMiddleware(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		deviceID := ""

		cookie, err := r.Cookie(deviceIDCookie)
		if err == nil {
			deviceID = cookie.Value
		}

		if deviceID == "" {
			deviceID = uuid.New().String()

			http.SetCookie(w, &http.Cookie{
				Name:     deviceIDCookie,
				Value:    deviceID,
				Path:     "/",
				MaxAge:   86400 * 365,
				HttpOnly: true,
				Secure:   SecureCookies(),
				SameSite: cookieSameSite(),
			})
		}
		ctx := context.WithValue(r.Context(), deviceIDKey, deviceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func StructuredLogger(log *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			start := time.Now()
			defer func() {
				log.Info("http_request",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("remote_addr", r.RemoteAddr),
					zap.String("user_agent", r.UserAgent()),
					zap.Int("status", ww.Status()),
					zap.Int("bytes", ww.BytesWritten()),
					zap.Duration("duration", time.Since(start)),
					zap.String("request_id", middleware.GetReqID(r.Context())),
				)
			}()

			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}

type tokenBucket struct {
	tokens   float64
	lastSeen time.Time
}

type rateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*tokenBucket

	rate  float64
	burst float64
}

func newRateLimiter(rate, burst float64) *rateLimiter {
	rl := &rateLimiter{
		buckets: make(map[string]*tokenBucket),
		rate:    rate,
		burst:   burst,
	}

	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			rl.evict(5 * time.Minute)
		}
	}()

	return rl
}

func (rl *rateLimiter) evict(idle time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	cutoff := time.Now().Add(-idle)
	for key, b := range rl.buckets {
		if b.lastSeen.Before(cutoff) {
			delete(rl.buckets, key)
		}
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.buckets[key]
	if !ok {
		rl.buckets[key] = &tokenBucket{tokens: rl.burst - 1, lastSeen: now}
		return true
	}

	b.tokens = min(rl.burst, b.tokens+now.Sub(b.lastSeen).Seconds()*rl.rate)
	b.lastSeen = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func clientKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

var defaultRateLimiter = newRateLimiter(20, 60)

func RateLimiterMiddleware(next http.Handler) http.Handler {
	return RateLimiterWith(defaultRateLimiter)(next)
}

func RateLimiterWith(rl *rateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !rl.allow(clientKey(r)) {
				w.Header().Set("Retry-After", "1")
				utils.ReturnErrorJSON(w, "Too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func NewTestRateLimiter(rate, burst float64) *rateLimiter {
	return &rateLimiter{
		buckets: make(map[string]*tokenBucket),
		rate:    rate,
		burst:   burst,
	}
}
