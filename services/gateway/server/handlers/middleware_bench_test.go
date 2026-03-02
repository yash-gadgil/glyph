package handlers

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/mock"
	"github.com/yash-gadgil/glyph/services/gateway/tests/mocks"
	authpb "github.com/yash-gadgil/glyph/services/gen/golang/auth"
)

func benchKey(b *testing.B) (*rsa.PrivateKey, *authpb.PublicKey) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		b.Fatal(err)
	}
	pub := &priv.PublicKey
	jwk := &authpb.PublicKey{
		Kid: "bench-kid",
		Kty: "RSA",
		Use: "sig",
		Alg: "RS256",
		N:   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
	}
	return priv, jwk
}

func BenchmarkAuthMiddlewareWarmCache(b *testing.B) {
	priv, jwk := benchKey(b)
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"userid": "bench-user",
		"exp":    time.Now().Add(time.Hour).Unix(),
	})
	tok.Header["kid"] = "bench-kid"
	signed, err := tok.SignedString(priv)
	if err != nil {
		b.Fatal(err)
	}

	mockAuth := new(mocks.MockAuthClient)
	mockAuth.On("GetPublicKeys", mock.Anything, mock.Anything).
		Return(&authpb.GetPublicKeysResponse{Keys: []*authpb.PublicKey{jwk}}, nil)

	cfg := NewTestConfig(mockAuth)
	handler := cfg.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	newReq := func() *http.Request {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.AddCookie(&http.Cookie{Name: "accessToken", Value: signed})
		return req
	}

	handler.ServeHTTP(httptest.NewRecorder(), newReq())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, newReq())
		if w.Code != http.StatusOK {
			b.Fatalf("unexpected status %d", w.Code)
		}
	}
}

func BenchmarkJWKSToRSAPublicKey(b *testing.B) {
	_, jwk := benchKey(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := jwksToRSAPublicKey(jwk.N, jwk.E); err != nil {
			b.Fatal(err)
		}
	}
}
