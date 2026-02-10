package db

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestCache(t *testing.T) (*Cache, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return &Cache{Rdb: rdb, log: zap.NewNop()}, mr
}

func TestPendingSignupRoundTrip(t *testing.T) {
	cache, mr := newTestCache(t)
	ctx := context.Background()

	err := cache.StorePendingSignup(ctx, "user@example.com", "Yash", "hash123", 30*time.Minute)
	require.NoError(t, err)

	got, err := cache.GetPendingSignup(ctx, "user@example.com")
	require.NoError(t, err)
	assert.Equal(t, "Yash", got.Name)
	assert.Equal(t, "hash123", got.PasswordHash)

	assert.Greater(t, mr.TTL("pending_signup:user@example.com"), time.Duration(0))

	require.NoError(t, cache.DeletePendingSignup(ctx, "user@example.com"))
	_, err = cache.GetPendingSignup(ctx, "user@example.com")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found or expired")
}

func TestGetPendingSignupExpires(t *testing.T) {
	cache, mr := newTestCache(t)
	ctx := context.Background()

	require.NoError(t, cache.StorePendingSignup(ctx, "user@example.com", "Yash", "hash123", time.Minute))
	mr.FastForward(2 * time.Minute)

	_, err := cache.GetPendingSignup(ctx, "user@example.com")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found or expired")
}

func TestGetPendingSignupCorruptPayload(t *testing.T) {
	cache, mr := newTestCache(t)
	require.NoError(t, mr.Set("pending_signup:bad@example.com", "{not-json"))

	_, err := cache.GetPendingSignup(context.Background(), "bad@example.com")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestVerificationEmailQueueRoundTrip(t *testing.T) {
	cache, _ := newTestCache(t)
	ctx := context.Background()

	require.NoError(t, cache.EnqueueVerificationEmail(ctx, "user@example.com", "tok-1"))

	payload, err := cache.DequeueVerificationEmail(ctx)
	require.NoError(t, err)

	var job struct {
		Email string `json:"email"`
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal([]byte(payload), &job))
	assert.Equal(t, "user@example.com", job.Email)
	assert.Equal(t, "tok-1", job.Token)
}

func TestPasswordResetQueueRoundTrip(t *testing.T) {
	cache, _ := newTestCache(t)
	ctx := context.Background()

	require.NoError(t, cache.EnqueuePasswordResetEmail(ctx, "user@example.com", "tok-2"))

	payload, err := cache.DequeuePasswordResetEmail(ctx)
	require.NoError(t, err)
	assert.Contains(t, payload, "tok-2")
}

func TestMoveVerificationToDLQ(t *testing.T) {
	cache, mr := newTestCache(t)
	ctx := context.Background()

	payload := `{"email":"user@example.com","token":"tok-1"}`
	require.NoError(t, cache.MoveVerificationToDLQ(ctx, payload, "send failed"))

	entries, err := mr.List("email_verification_dlq")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Contains(t, entries[0], "send failed")
	assert.Contains(t, entries[0], "tok-1")
}

func TestQueueOrderingIsFIFO(t *testing.T) {
	cache, _ := newTestCache(t)
	ctx := context.Background()

	require.NoError(t, cache.EnqueueVerificationEmail(ctx, "first@example.com", "t1"))
	require.NoError(t, cache.EnqueueVerificationEmail(ctx, "second@example.com", "t2"))

	first, err := cache.DequeueVerificationEmail(ctx)
	require.NoError(t, err)
	assert.Contains(t, first, "first@example.com")

	second, err := cache.DequeueVerificationEmail(ctx)
	require.NoError(t, err)
	assert.Contains(t, second, "second@example.com")
}

func TestNilRedisDegradedMode(t *testing.T) {
	cache := &Cache{Rdb: nil, log: zap.NewNop()}
	ctx := context.Background()

	assert.Error(t, cache.StorePendingSignup(ctx, "a@b.co", "n", "h", time.Minute))
	_, err := cache.GetPendingSignup(ctx, "a@b.co")
	assert.Error(t, err)
	assert.NoError(t, cache.DeletePendingSignup(ctx, "a@b.co"))
	assert.NoError(t, cache.MoveVerificationToDLQ(ctx, "{}", "err"))
	assert.Error(t, cache.EnqueueVerificationEmail(ctx, "a@b.co", "t"))
	assert.Error(t, cache.EnqueuePasswordResetEmail(ctx, "a@b.co", "t"))
	_, err = cache.DequeueVerificationEmail(ctx)
	assert.Error(t, err)
	_, err = cache.DequeuePasswordResetEmail(ctx)
	assert.Error(t, err)
}

func TestInitCacheWithoutAddrReturnsDegradedCache(t *testing.T) {
	t.Setenv("CACHE_ADDR", "")
	cache := InitCache(context.Background(), zap.NewNop())
	require.NotNil(t, cache)
	assert.Nil(t, cache.Rdb)
}

func TestInitCacheWithUnreachableAddrReturnsDegradedCache(t *testing.T) {
	t.Setenv("CACHE_ADDR", "127.0.0.1:1")
	t.Setenv("CACHE_PASSWORD", "")
	cache := InitCache(context.Background(), zap.NewNop())
	require.NotNil(t, cache)
	assert.Nil(t, cache.Rdb)
}

func TestInitCacheConnectsToRealRedis(t *testing.T) {
	mr := miniredis.RunT(t)
	t.Setenv("CACHE_ADDR", mr.Addr())
	t.Setenv("CACHE_PASSWORD", "")

	cache := InitCache(context.Background(), zap.NewNop())
	require.NotNil(t, cache)
	require.NotNil(t, cache.Rdb)
	assert.NoError(t, cache.StorePendingSignup(context.Background(), "a@b.co", "n", "h", time.Minute))
}
