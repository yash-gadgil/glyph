package handlers

import (
	"github.com/yash-gadgil/glyph/services/auth/db"
	"github.com/yash-gadgil/glyph/services/auth/utils"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

func NewTestAuthHandler(cfg *oauth2.Config, cache *db.Cache, keyStore *utils.KeyStore, client userpb.AccountServiceClient, log *zap.Logger) *AuthHandler {
	return &AuthHandler{
		googleConfig: cfg,
		cache:        cache,
		keyStore:     keyStore,
		userClient:   client,
		log:          log,
	}
}
