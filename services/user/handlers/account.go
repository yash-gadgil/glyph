package handlers

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/yash-gadgil/glyph/pkg/logger"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	db "github.com/yash-gadgil/glyph/services/user/db/gen"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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

func (s *AccountHandler) SigninUser(ctx context.Context, req *userpb.SigninUserInfo) (*userpb.UserSpecifier, error) {
	log := logger.WithContextFields(ctx, s.log).With(logger.Action("signin_user"))

	res, err := s.q.GetUserPassword(ctx, req.Email)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	if res.PasswordHash != nil && req.Password != nil {
		if err := bcrypt.CompareHashAndPassword([]byte(*res.PasswordHash), []byte(*req.Password)); err != nil {
			return nil, status.Errorf(codes.PermissionDenied, "incorrect password")
		}
	}

	if res.PasswordHash == nil && req.Password != nil {
		return nil, status.Errorf(codes.PermissionDenied, "this account uses Google sign-in")
	}

	if err := s.q.EnsureAccount(ctx, res.ID); err != nil {
		log.Error("ensure_account_failed", zap.Error(err))
	}

	log.Info("user_signed_in", logger.KV("user_id", res.ID.String()))

	return &userpb.UserSpecifier{UserId: res.ID.String()}, nil
}

func (s *AccountHandler) CheckEmailAvailability(ctx context.Context, req *userpb.CheckEmailRequest) (*userpb.CheckEmailResponse, error) {
	present, err := s.q.CheckEmailAvailability(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	return &userpb.CheckEmailResponse{Available: !present}, nil
}

func (s *AccountHandler) UpdatePasswordByEmail(ctx context.Context, req *userpb.UpdatePasswordRequest) (*emptypb.Empty, error) {
	if err := s.q.UpdateUserPasswordByEmail(ctx, db.UpdateUserPasswordByEmailParams{
		PasswordHash: &req.PasswordHash,
		Email:        req.Email,
	}); err != nil {
		s.log.Error("password_update_failed", logger.Action("update_password"), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update password")
	}

	return &emptypb.Empty{}, nil
}

func (s *AccountHandler) GetProfile(ctx context.Context, req *userpb.UserSpecifier) (*userpb.Profile, error) {
	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	row, err := s.q.GetUserById(ctx, userUUID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	return &userpb.Profile{
		UserId:   row.ID.String(),
		Email:    row.Email,
		UserName: row.UserName,
	}, nil
}
