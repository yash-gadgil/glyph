package logger

import (
	"go.uber.org/zap"
)

func New(serviceName string) *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.Encoding = "json"

	l, err := cfg.Build()
	if err != nil {
		l = zap.NewNop()
	}

	return l.With(zap.String("service", serviceName))
}
