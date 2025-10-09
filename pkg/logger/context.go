package logger

import (
	"context"

	"go.uber.org/zap"
)

type contextKey string

const (
	CtxRequestID contextKey = "request_id"
	CtxUserID    contextKey = "user_id"
)

func WithContextFields(ctx context.Context, log *zap.Logger) *zap.Logger {
	if reqID, ok := ctx.Value(CtxRequestID).(string); ok && reqID != "" {
		log = log.With(zap.String("request_id", reqID))
	}

	if userID, ok := ctx.Value(CtxUserID).(string); ok && userID != "" {
		log = log.With(zap.String("user_id", userID))
	}

	return log
}
