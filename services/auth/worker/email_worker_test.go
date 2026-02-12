package worker

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yash-gadgil/glyph/services/auth/db"
	"go.uber.org/zap"
)

type sentEmail struct {
	to, subject, body string
}

type fakeSender struct {
	mu    sync.Mutex
	sent  []sentEmail
	fail  bool
	calls int
}

func (f *fakeSender) send(to, subject, body string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	if f.fail {
		return fmt.Errorf("smtp unavailable")
	}
	f.sent = append(f.sent, sentEmail{to, subject, body})
	return nil
}

func withFakeSender(t *testing.T, fail bool) *fakeSender {
	t.Helper()
	fake := &fakeSender{fail: fail}
	origSend := sendEmail
	origDelay := retryBaseDelay
	sendEmail = fake.send
	retryBaseDelay = 0
	t.Cleanup(func() {
		sendEmail = origSend
		retryBaseDelay = origDelay
	})
	return fake
}

func newTestCache(t *testing.T) (*db.Cache, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return &db.Cache{Rdb: rdb}, mr
}

func TestProcessNextEmailSendsVerificationLink(t *testing.T) {
	fake := withFakeSender(t, false)
	cache, _ := newTestCache(t)
	t.Setenv("EMAIL_VERIFICATION_URL", "https://glyph.example.com")

	require.NoError(t, cache.EnqueueVerificationEmail(context.Background(), "user@example.com", "tok-123"))
	require.NoError(t, processNextEmail(context.Background(), cache, zap.NewNop()))

	require.Len(t, fake.sent, 1)
	assert.Equal(t, "user@example.com", fake.sent[0].to)
	assert.Equal(t, "Verify your email", fake.sent[0].subject)
	assert.Contains(t, fake.sent[0].body, "https://glyph.example.com/auth/verify?token=tok-123")
}

func TestProcessNextEmailEmptyQueueIsNoop(t *testing.T) {
	fake := withFakeSender(t, false)
	cache, mr := newTestCache(t)

	mr.Close()
	err := processNextEmail(context.Background(), cache, zap.NewNop())
	assert.Error(t, err)
	assert.Empty(t, fake.sent)
}

func TestProcessNextEmailMovesBadPayloadToDLQ(t *testing.T) {
	fake := withFakeSender(t, false)
	cache, mr := newTestCache(t)

	require.NoError(t, cache.Rdb.RPush(context.Background(), "email_verification_queue", "{not-json").Err())
	require.NoError(t, processNextEmail(context.Background(), cache, zap.NewNop()))

	assert.Empty(t, fake.sent)
	entries, err := mr.List("email_verification_dlq")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Contains(t, entries[0], "invalid JSON")
}

func TestProcessNextEmailRetriesThenDLQs(t *testing.T) {
	fake := withFakeSender(t, true)
	cache, mr := newTestCache(t)

	require.NoError(t, cache.EnqueueVerificationEmail(context.Background(), "user@example.com", "tok-123"))
	require.NoError(t, processNextEmail(context.Background(), cache, zap.NewNop()))

	assert.Equal(t, 4, fake.calls, "should retry the send 4 times")
	entries, err := mr.List("email_verification_dlq")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Contains(t, entries[0], "send failed")
}

func TestProcessNextPasswordResetSendsResetLink(t *testing.T) {
	fake := withFakeSender(t, false)
	cache, _ := newTestCache(t)
	t.Setenv("FRONTEND_URL", "https://app.glyph.example.com")

	require.NoError(t, cache.EnqueuePasswordResetEmail(context.Background(), "user@example.com", "tok-456"))
	require.NoError(t, processNextPasswordReset(context.Background(), cache, zap.NewNop()))

	require.Len(t, fake.sent, 1)
	assert.Equal(t, "Reset your password", fake.sent[0].subject)
	assert.Contains(t, fake.sent[0].body, "https://app.glyph.example.com/reset-password?token=tok-456")
}

func TestProcessNextPasswordResetSkipsBadPayload(t *testing.T) {
	fake := withFakeSender(t, false)
	cache, _ := newTestCache(t)

	require.NoError(t, cache.Rdb.RPush(context.Background(), "password_reset_queue", "{not-json").Err())
	require.NoError(t, processNextPasswordReset(context.Background(), cache, zap.NewNop()))
	assert.Empty(t, fake.sent)
}

func TestStartEmailWorkerDisabledWithoutRedis(t *testing.T) {
	stop := StartEmailWorker(&db.Cache{Rdb: nil}, zap.NewNop())
	require.NotNil(t, stop)
	close(stop)
}

func TestStartEmailWorkerProcessesQueueEndToEnd(t *testing.T) {
	fake := withFakeSender(t, false)
	cache, _ := newTestCache(t)
	t.Setenv("EMAIL_VERIFICATION_URL", "https://glyph.example.com")

	require.NoError(t, cache.EnqueueVerificationEmail(context.Background(), "worker@example.com", "tok-789"))

	stop := StartEmailWorker(cache, zap.NewNop())
	defer close(stop)

	require.Eventually(t, func() bool {
		fake.mu.Lock()
		defer fake.mu.Unlock()
		return len(fake.sent) == 1
	}, 5*time.Second, 50*time.Millisecond)

	fake.mu.Lock()
	defer fake.mu.Unlock()
	assert.Equal(t, "worker@example.com", fake.sent[0].to)
}
