package handlers

import (
	"github.com/yash-gadgil/glyph/services/auth/db"
	"github.com/yash-gadgil/glyph/services/auth/types"
	"github.com/yash-gadgil/glyph/services/auth/utils"
	authpb "github.com/yash-gadgil/glyph/services/gen/golang/auth"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
)

type AuthHandler struct {
	authpb.UnimplementedAuthServiceServer

	googleConfig *oauth2.Config
	AddrConfig   *types.AddrConfig

	log *zap.Logger

	userConn   *grpc.ClientConn
	userClient userpb.AccountServiceClient

	cache    *db.Cache
	keyStore *utils.KeyStore
}

func NewAuthHandler(cfg *oauth2.Config, acfg *types.AddrConfig, cache *db.Cache, keyStore *utils.KeyStore, log *zap.Logger) *AuthHandler {
	h := &AuthHandler{
		googleConfig: cfg,
		AddrConfig:   acfg,
		cache:        cache,
		keyStore:     keyStore,
		log:          log,
	}
	if acfg != nil && acfg.UserSvcAddr != "" {
		conn := utils.GetGrpcClient(acfg.UserSvcAddr)
		if conn != nil {
			h.userConn = conn
			h.userClient = userpb.NewAccountServiceClient(conn)
		}
	}
	return h
}

func (h *AuthHandler) Close() error {
	if h.userConn != nil {
		return h.userConn.Close()
	}
	return nil
}

func Register(grpcServer *grpc.Server, h *AuthHandler) {
	authpb.RegisterAuthServiceServer(grpcServer, h)
}
