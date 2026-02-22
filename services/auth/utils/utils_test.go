package utils

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newTestKeyStore(t *testing.T) *KeyStore {
	t.Helper()
	ks, err := NewKeyStore(zap.NewNop())
	require.NoError(t, err)
	return ks
}

func TestValidateEmail(t *testing.T) {
	cases := []struct {
		name    string
		email   string
		wantErr string
	}{
		{"valid simple", "user@example.com", ""},
		{"valid with plus", "user+tag@example.com", ""},
		{"valid subdomain", "user@mail.example.co", ""},
		{"empty", "", "email is required"},
		{"missing at", "userexample.com", "format is invalid"},
		{"missing domain dot", "user@examplecom", "format is invalid"},
		{"missing local part", "@example.com", "format is invalid"},
		{"spaces trimmed ok", "  user@example.com  ", ""},
		{"too long", strings.Repeat("a", 250) + "@example.com", "too long"},
		{"local part too long", strings.Repeat("a", 65) + "@example.com", "local part"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateEmail(tc.email)
			if tc.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		})
	}
}

func TestValidatePassword(t *testing.T) {
	cases := []struct {
		name    string
		pw      string
		wantErr string
	}{
		{"valid", "Passw0rd!", ""},
		{"valid all classes", "aB3$efgh", ""},
		{"too short", "aB3$efg", "at least 8"},
		{"too long", "aB3$efghaB3$efghX", "no more than 16"},
		{"no lowercase", "PASSW0RD!", "lowercase"},
		{"no uppercase", "passw0rd!", "uppercase"},
		{"no digit", "Password!", "digit"},
		{"no special", "Passw0rdd", "special"},
		{"contains space", "Pass w0rd!", "spaces"},
		{"contains underscore", "Pass_w0rd!", "invalid character '_'"},
		{"non-ascii rejected", "Pässw0rd!", "invalid character"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePassword(tc.pw)
			if tc.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		})
	}
}

func TestValidateName(t *testing.T) {
	assert.NoError(t, ValidateName("Yash"))
	assert.NoError(t, ValidateName("y"))
	assert.NoError(t, ValidateName(strings.Repeat("y", 20)))
	assert.Error(t, ValidateName(""))
	assert.Error(t, ValidateName(strings.Repeat("y", 21)))
}

func TestCreateAndVerifyTokenRoundTrip(t *testing.T) {
	ks := newTestKeyStore(t)

	token, err := CreateToken("user-123", time.Now().Add(time.Hour), ks.GetCurrentKey())
	require.NoError(t, err)
	require.NotEmpty(t, token)

	userID, err := VerifyToken(token, ks)
	require.NoError(t, err)
	assert.Equal(t, "user-123", userID)

	viaGet, err := GetUserIDFromToken(token, ks)
	require.NoError(t, err)
	assert.Equal(t, "user-123", viaGet)
}

func TestExpiredTokenIsRejected(t *testing.T) {
	ks := newTestKeyStore(t)

	token, err := CreateToken("user-123", time.Now().Add(-time.Minute), ks.GetCurrentKey())
	require.NoError(t, err)

	_, err = VerifyToken(token, ks)
	assert.Error(t, err)

	_, err = ParseTokenClaims(token, ks)
	assert.Error(t, err)

	_, err = GetUserIDFromToken(token, ks)
	assert.Error(t, err)
}

func TestTokenSignedByUnknownKeyIsRejected(t *testing.T) {
	signer := newTestKeyStore(t)
	verifier := newTestKeyStore(t)

	token, err := CreateToken("user-123", time.Now().Add(time.Hour), signer.GetCurrentKey())
	require.NoError(t, err)

	_, err = VerifyToken(token, verifier)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "public key")
}

func TestTokenWithWrongSigningMethodIsRejected(t *testing.T) {
	ks := newTestKeyStore(t)

	hs := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userid": "user-123",
		"exp":    time.Now().Add(time.Hour).Unix(),
	})
	hs.Header["kid"] = ks.GetCurrentKey().ID
	tokenStr, err := hs.SignedString([]byte("secret"))
	require.NoError(t, err)

	_, err = VerifyToken(tokenStr, ks)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "signing method")
}

func TestTokenWithoutKidIsRejected(t *testing.T) {
	ks := newTestKeyStore(t)

	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"userid": "user-123",
		"exp":    time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := tok.SignedString(ks.GetCurrentKey().PrivateKey)
	require.NoError(t, err)

	_, err = VerifyToken(tokenStr, ks)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "kid")
}

func TestTokenWithoutUserIDClaim(t *testing.T) {
	ks := newTestKeyStore(t)

	token, err := CreateTokenWithClaims(map[string]any{"email": "a@b.co"}, time.Now().Add(time.Hour), ks.GetCurrentKey())
	require.NoError(t, err)

	_, err = VerifyToken(token, ks)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "userid")
}

func TestParseTokenClaimsReturnsCustomClaims(t *testing.T) {
	ks := newTestKeyStore(t)

	token, err := CreateTokenWithClaims(map[string]any{
		"email":   "user@example.com",
		"purpose": "password_reset",
	}, time.Now().Add(time.Hour), ks.GetCurrentKey())
	require.NoError(t, err)

	claims, err := ParseTokenClaims(token, ks)
	require.NoError(t, err)
	assert.Equal(t, "user@example.com", claims["email"])
	assert.Equal(t, "password_reset", claims["purpose"])
	assert.NotNil(t, claims["exp"])
	assert.NotNil(t, claims["iat"])
}

func TestGarbageTokenIsRejected(t *testing.T) {
	ks := newTestKeyStore(t)
	_, err := VerifyToken("not-a-jwt", ks)
	assert.Error(t, err)
	_, err = ParseTokenClaims("", ks)
	assert.Error(t, err)
}
