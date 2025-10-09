package logger

import "go.uber.org/zap"

func Action(a string) zap.Field {
	return zap.String("action", a)
}

func Stage(s string) zap.Field {
	return zap.String("stage", s)
}

func KV(k, v string) zap.Field {
	return zap.String(k, v)
}

func ErrStr(e string) zap.Field {
	return zap.String("error", e)
}
