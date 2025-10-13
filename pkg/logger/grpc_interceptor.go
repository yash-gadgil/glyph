package logger

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func UnaryServerInterceptor(base *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {

		start := time.Now()
		method := info.FullMethod

		log := WithContextFields(ctx, base)

		resp, err = handler(ctx, req)
		latency := time.Since(start)

		fields := []zap.Field{
			zap.String("method", method),
			zap.Duration("latency_ms", latency),
			Action("grpc_call"),
			Stage("handler"),
		}

		if err != nil {
			log.Error("grpc_error", append(fields, zap.Error(err))...)
		} else {
			log.Info("grpc_request_completed", fields...)
		}

		return resp, err
	}
}
