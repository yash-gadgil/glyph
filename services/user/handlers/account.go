package handlers

import (
	"context"
	"database/sql"

	"github.com/yash-gadgil/glyph/pkg/logger"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	db "github.com/yash-gadgil/glyph/services/user/db/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AccountHandler struct {
	userpb.UnimplementedAccountServiceServer
	db  *sql.DB
	q   *db.Queries
	log *zap.Logger
}

func NewAccountHandler(sdb *sql.DB, log *zap.Logger) *AccountHandler {
	return &AccountHandler{db: sdb, q: db.New(sdb), log: log}
}

func (s *AccountHandler) SignupUser(ctx context.Context, req *userpb.SignupUserInfo) (*userpb.UserSpecifier, error) {
	log := logger.WithContextFields(ctx, s.log).With(logger.Action("signup_user"))

	id, err := s.q.CreateUser(ctx, db.CreateUserParams{
		UserName:     req.UserName,
		Email:        req.Email,
		PasswordHash: req.Password,
	})
	if err != nil {
		log.Error("create_user_failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Error creating user")
	}

	if _, err := s.q.CreateAccount(ctx, id); err != nil {
		log.Error("create_account_failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Error creating user")
	}

	log.Info("user_registered", logger.KV("user_id", id.String()))

	return &userpb.UserSpecifier{UserId: id.String()}, nil
}
