package handlers_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yash-gadgil/glyph/services/gateway/server/handlers"
	"github.com/yash-gadgil/glyph/services/gateway/tests/mocks"
	authpb "github.com/yash-gadgil/glyph/services/gen/golang/auth"
)

func signToken(t *testing.T, key *rsa.PrivateKey, kid, userID string) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{"userid": userID})
	tok.Header["kid"] = kid
	s, err := tok.SignedString(key)
	require.NoError(t, err)
	return s
}

func publicKeyMessage(kid string, pub *rsa.PublicKey) *authpb.PublicKey {
	return &authpb.PublicKey{
		Kid: kid,
		N:   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
	}
}

func TestAuthMiddlewareRejectsMissingCookie(t *testing.T) {
	cfg := handlers.NewTestConfig(new(mocks.MockAuthClient))

	called := false
	h := cfg.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/account", nil))

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, called)
}

func TestAuthMiddlewareAcceptsValidToken(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	authClient := new(mocks.MockAuthClient)
	authClient.On("GetPublicKeys", mock.Anything, mock.Anything).Return(&authpb.GetPublicKeysResponse{
		Keys: []*authpb.PublicKey{publicKeyMessage("kid-1", &key.PublicKey)},
	}, nil)

	cfg := handlers.NewTestConfig(authClient)

	called := false
	h := cfg.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/account", nil)
	req.AddCookie(&http.Cookie{Name: "accessToken", Value: signToken(t, key, "kid-1", "user-42")})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, called)
	authClient.AssertExpectations(t)
}

func TestAuthMiddlewareRejectsUnknownKID(t *testing.T) {
	authClient := new(mocks.MockAuthClient)
	authClient.On("GetPublicKeys", mock.Anything, mock.Anything).Return(&authpb.GetPublicKeysResponse{}, nil).Maybe()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	cfg := handlers.NewTestConfig(authClient)

	h := cfg.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/account", nil)
	req.AddCookie(&http.Cookie{Name: "accessToken", Value: signToken(t, key, "missing-kid", "user-7")})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestDeviceIDMiddlewareSetsCookie(t *testing.T) {
	cfg := handlers.NewTestConfig(new(mocks.MockAuthClient))

	h := cfg.DeviceIDMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	var found bool
	for _, c := range rec.Result().Cookies() {
		if c.Name == "device_id" && c.Value != "" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestRateLimiterBlocksBurst(t *testing.T) {
	rl := handlers.NewTestRateLimiter(1, 2)

	h := handlers.RateLimiterWith(rl)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	var lastCode int
	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		h.ServeHTTP(rec, req)
		lastCode = rec.Code
	}

	assert.Equal(t, http.StatusTooManyRequests, lastCode)
}
