package handlers

import (
	"context"
	"net/http"
)

func WithUserID(r *http.Request, userID string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), userIDKey, userID))
}
