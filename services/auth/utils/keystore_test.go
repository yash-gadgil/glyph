package utils

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewKeyStoreGeneratesInitialKey(t *testing.T) {
	ks := newTestKeyStore(t)

	key := ks.GetCurrentKey()
	require.NotNil(t, key)
	assert.NotEmpty(t, key.ID)
	assert.NotNil(t, key.PrivateKey)
	assert.NotNil(t, key.PublicKey)
	assert.True(t, key.ExpiresAt.After(time.Now()))
}

func TestGetPublicKeyForCurrentAndUnknownKid(t *testing.T) {
	ks := newTestKeyStore(t)

	current := ks.GetCurrentKey()
	pub, err := ks.GetPublicKey(current.ID)
	require.NoError(t, err)
	assert.Equal(t, current.PublicKey, pub)

	_, err = ks.GetPublicKey("does-not-exist")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does-not-exist")
}

func TestGetAllPublicKeysIncludesCurrent(t *testing.T) {
	ks := newTestKeyStore(t)

	keys := ks.GetAllPublicKeys()
	require.Len(t, keys, 1)
	assert.Equal(t, ks.GetCurrentKey().ID, keys[0].ID)
}

func TestRotationKeepsOldKeyVerifiable(t *testing.T) {
	ks := newTestKeyStore(t)
	oldKey := ks.GetCurrentKey()

	token, err := CreateToken("user-123", time.Now().Add(time.Hour), oldKey)
	require.NoError(t, err)

	stop := ks.StartRotationScheduler(10 * time.Millisecond)
	defer close(stop)

	require.Eventually(t, func() bool {
		return ks.GetCurrentKey().ID != oldKey.ID
	}, 2*time.Second, 10*time.Millisecond, "expected key rotation to happen")

	pub, err := ks.GetPublicKey(oldKey.ID)
	require.NoError(t, err)
	assert.Equal(t, oldKey.PublicKey, pub)

	userID, err := VerifyToken(token, ks)
	require.NoError(t, err)
	assert.Equal(t, "user-123", userID)

	assert.GreaterOrEqual(t, len(ks.GetAllPublicKeys()), 2)
}

func TestExportPublicKeyPEM(t *testing.T) {
	ks := newTestKeyStore(t)

	pem, err := ExportPublicKeyPEM(ks.GetCurrentKey().PublicKey)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(pem, "-----BEGIN RSA PUBLIC KEY-----"))
	assert.Contains(t, pem, "-----END RSA PUBLIC KEY-----")
}

func TestPublicKeyToJWK(t *testing.T) {
	ks := newTestKeyStore(t)
	key := ks.GetCurrentKey()

	jwk := PublicKeyToJWK(key.ID, key.PublicKey)
	assert.Equal(t, "RSA", jwk["kty"])
	assert.Equal(t, "sig", jwk["use"])
	assert.Equal(t, "RS256", jwk["alg"])
	assert.Equal(t, key.ID, jwk["kid"])
	assert.NotEmpty(t, jwk["n"])
	assert.NotEmpty(t, jwk["e"])
	assert.NotContains(t, jwk["n"], "=")
	assert.NotContains(t, jwk["e"], "=")
}

func TestKeyStoreNopLoggerSafety(t *testing.T) {
	ks, err := NewKeyStore(zap.NewNop())
	require.NoError(t, err)
	stop := ks.StartRotationScheduler(time.Hour)
	close(stop)
}
