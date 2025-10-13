package logger

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc"
)

func observed() (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zapcore.DebugLevel)
	return zap.New(core), logs
}

func fieldMap(entry observer.LoggedEntry) map[string]any {
	m := map[string]any{}
	for _, f := range entry.Context {
		m[f.Key] = f.String
		if f.Type == zapcore.Int64Type {
			m[f.Key] = f.Integer
		}
	}
	return m
}

func TestNewReturnsUsableLogger(t *testing.T) {
	log := New("test-service")
	require.NotNil(t, log)
	entry := log.Check(zapcore.InfoLevel, "probe")
	require.NotNil(t, entry)
}

func TestEventFieldHelpers(t *testing.T) {
	assert.Equal(t, zap.String("action", "signup"), Action("signup"))
	assert.Equal(t, zap.String("stage", "validate"), Stage("validate"))
	assert.Equal(t, zap.String("k", "v"), KV("k", "v"))
	assert.Equal(t, zap.String("error", "boom"), ErrStr("boom"))
}

func TestWithContextFieldsAddsKnownFields(t *testing.T) {
	log, logs := observed()

	ctx := context.WithValue(context.Background(), CtxRequestID, "req-1")
	ctx = context.WithValue(ctx, CtxUserID, "user-1")

	WithContextFields(ctx, log).Info("hello")

	require.Equal(t, 1, logs.Len())
	fields := fieldMap(logs.All()[0])
	assert.Equal(t, "req-1", fields["request_id"])
	assert.Equal(t, "user-1", fields["user_id"])
}

func TestWithContextFieldsSkipsMissingOrEmptyValues(t *testing.T) {
	log, logs := observed()

	ctx := context.WithValue(context.Background(), CtxRequestID, "")
	WithContextFields(ctx, log).Info("hello")

	require.Equal(t, 1, logs.Len())
	fields := fieldMap(logs.All()[0])
	assert.NotContains(t, fields, "request_id")
	assert.NotContains(t, fields, "user_id")
}

func TestWithContextFieldsIgnoresWrongTypes(t *testing.T) {
	log, logs := observed()

	ctx := context.WithValue(context.Background(), CtxRequestID, 42)
	WithContextFields(ctx, log).Info("hello")

	fields := fieldMap(logs.All()[0])
	assert.NotContains(t, fields, "request_id")
}

func TestUnaryInterceptorLogsSuccess(t *testing.T) {
	log, logs := observed()
	interceptor := UnaryServerInterceptor(log)

	ctx := context.WithValue(context.Background(), CtxUserID, "user-9")
	resp, err := interceptor(ctx, "request",
		&grpc.UnaryServerInfo{FullMethod: "/auth.AuthService/Signin"},
		func(ctx context.Context, req any) (any, error) {
			return "response", nil
		})

	require.NoError(t, err)
	assert.Equal(t, "response", resp)

	require.Equal(t, 1, logs.Len())
	entry := logs.All()[0]
	assert.Equal(t, "grpc_request_completed", entry.Message)
	assert.Equal(t, zapcore.InfoLevel, entry.Level)

	fields := fieldMap(entry)
	assert.Equal(t, "/auth.AuthService/Signin", fields["method"])
	assert.Equal(t, "user-9", fields["user_id"])
	assert.Equal(t, "grpc_call", fields["action"])
}

func TestUnaryInterceptorLogsErrorAndPropagates(t *testing.T) {
	log, logs := observed()
	interceptor := UnaryServerInterceptor(log)

	boom := fmt.Errorf("boom")
	resp, err := interceptor(context.Background(), "request",
		&grpc.UnaryServerInfo{FullMethod: "/m"},
		func(ctx context.Context, req any) (any, error) {
			return nil, boom
		})

	assert.Nil(t, resp)
	assert.Same(t, boom, err)

	require.Equal(t, 1, logs.Len())
	entry := logs.All()[0]
	assert.Equal(t, "grpc_error", entry.Message)
	assert.Equal(t, zapcore.ErrorLevel, entry.Level)
}

func TestHTTPRequestLoggerLogsRequest(t *testing.T) {
	log, logs := observed()

	handler := HTTPRequestLogger(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("short and stout"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	req = req.WithContext(context.WithValue(req.Context(), CtxRequestID, "req-7"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusTeapot, rec.Code)
	require.Equal(t, 1, logs.Len())

	entry := logs.All()[0]
	assert.Equal(t, "http_request", entry.Message)
	fields := fieldMap(entry)
	assert.Equal(t, "GET", fields["method"])
	assert.Equal(t, "/api/v1/orders", fields["path"])
	assert.Equal(t, int64(http.StatusTeapot), fields["status"])
	assert.Equal(t, int64(len("short and stout")), fields["bytes"])
	assert.Equal(t, "req-7", fields["request_id"])
}

func TestHTTPRequestLoggerPassesThroughBody(t *testing.T) {
	log, _ := observed()

	handler := HTTPRequestLogger(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/x", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())
}
