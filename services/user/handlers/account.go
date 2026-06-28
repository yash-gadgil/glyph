package handlers

import (
	"context"
	"database/sql"
	"errors"

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

const (
	SideBuy  int16 = 0
	SideSell int16 = 1
)

type UserEventPublisher interface {
	PublishUserDeleted(ctx context.Context, userID string) error
}

type AccountHandler struct {
	userpb.UnimplementedAccountServiceServer
	db  *sql.DB
	q   *db.Queries
	pub UserEventPublisher
	log *zap.Logger
}

func NewAccountHandler(sdb *sql.DB, pub UserEventPublisher, log *zap.Logger) *AccountHandler {
	return &AccountHandler{db: sdb, q: db.New(sdb), pub: pub, log: log}
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

func (s *AccountHandler) AddFunds(ctx context.Context, req *userpb.AddFundsRequest) (*userpb.AddFundsResponse, error) {
	return nil, status.Errorf(codes.PermissionDenied, "paper accounts have a fixed starting balance, reset the account instead")
}

func (s *AccountHandler) ResetAccount(ctx context.Context, req *userpb.UserSpecifier) (*emptypb.Empty, error) {
	log := logger.WithContextFields(ctx, s.log).With(logger.Action("reset_account"))

	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reset account")
	}
	defer tx.Rollback()

	qtx := s.q.WithTx(tx)
	if err := qtx.EnsureAccount(ctx, userUUID); err != nil {
		log.Error("reset_ensure_account_failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to reset account")
	}
	if err := qtx.ResetAccountBalances(ctx, userUUID); err != nil {
		log.Error("reset_balances_failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to reset account")
	}
	if err := qtx.DeletePositionsForUser(ctx, userUUID); err != nil {
		log.Error("reset_positions_failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to reset account")
	}
	if err := qtx.DeleteReservationsForUser(ctx, userUUID); err != nil {
		log.Error("reset_reservations_failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to reset account")
	}
	if err := qtx.DeleteSettlementsForUser(ctx, userUUID); err != nil {
		log.Error("reset_settlements_failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to reset account")
	}
	if err := tx.Commit(); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reset account")
	}

	log.Info("account_reset", logger.KV("user_id", req.UserId))
	return &emptypb.Empty{}, nil
}

func (s *AccountHandler) DeleteAccount(ctx context.Context, req *userpb.UserSpecifier) (*emptypb.Empty, error) {
	log := logger.WithContextFields(ctx, s.log).With(logger.Action("delete_account"))

	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	rows, err := s.q.DeleteUser(ctx, userUUID)
	if err != nil {
		log.Error("delete_user_failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to delete account")
	}
	if rows == 0 {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	if s.pub != nil {
		if err := s.pub.PublishUserDeleted(ctx, req.UserId); err != nil {
			log.Error("publish_user_deleted_failed", logger.KV("user_id", req.UserId), zap.Error(err))
		}
	}

	log.Info("account_deleted", logger.KV("user_id", req.UserId))
	return &emptypb.Empty{}, nil
}

func (s *AccountHandler) ReserveForOrder(ctx context.Context, req *userpb.ReserveForOrderRequest) (*emptypb.Empty, error) {
	log := logger.WithContextFields(ctx, s.log).With(logger.Action("reserve_for_order"))

	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}
	orderUUID, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid order ID")
	}
	if req.Qty <= 0 || req.CentsPerShare <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "qty and cents_per_share must be positive")
	}
	if side := int16(req.Side); side != SideBuy && side != SideSell {
		return nil, status.Errorf(codes.InvalidArgument, "invalid side %d", req.Side)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reserve")
	}
	defer tx.Rollback()

	qtx := s.q.WithTx(tx)

	switch int16(req.Side) {
	case SideBuy:
		hold := req.Qty * req.CentsPerShare
		if _, err := qtx.ReserveCash(ctx, db.ReserveCashParams{
			UserID:       userUUID,
			ReservedCash: hold,
		}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, status.Errorf(codes.FailedPrecondition, "insufficient buying power")
			}
			log.Error("reserve_cash_failed", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "failed to reserve")
		}
	case SideSell:
		if _, err := qtx.ReserveShares(ctx, db.ReserveSharesParams{
			UserID:      userUUID,
			Symbol:      req.Symbol,
			ReservedQty: req.Qty,
		}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, status.Errorf(codes.FailedPrecondition, "insufficient shares to sell")
			}
			log.Error("reserve_shares_failed", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "failed to reserve")
		}
	}

	if err := qtx.CreateReservation(ctx, db.CreateReservationParams{
		OrderID:       orderUUID,
		UserID:        userUUID,
		Symbol:        req.Symbol,
		Side:          int16(req.Side),
		Qty:           req.Qty,
		CentsPerShare: req.CentsPerShare,
	}); err != nil {
		log.Error("create_reservation_failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to reserve")
	}

	if err := tx.Commit(); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reserve")
	}

	return &emptypb.Empty{}, nil
}

func (s *AccountHandler) ReleaseForOrder(ctx context.Context, req *userpb.ReleaseForOrderRequest) (*emptypb.Empty, error) {
	log := logger.WithContextFields(ctx, s.log).With(logger.Action("release_for_order"))

	orderUUID, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid order ID")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to release")
	}
	defer tx.Rollback()

	qtx := s.q.WithTx(tx)

	res, err := qtx.GetReservationForUpdate(ctx, orderUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &emptypb.Empty{}, nil
		}
		log.Error("get_reservation_failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to release")
	}

	if res.RemainingQty > 0 {
		switch res.Side {
		case SideBuy:
			if err := qtx.ReleaseCash(ctx, db.ReleaseCashParams{
				UserID:       res.UserID,
				ReservedCash: res.RemainingQty * res.CentsPerShare,
			}); err != nil {
				log.Error("release_cash_failed", zap.Error(err))
				return nil, status.Errorf(codes.Internal, "failed to release")
			}
		case SideSell:
			if err := qtx.ReleaseShares(ctx, db.ReleaseSharesParams{
				UserID:      res.UserID,
				Symbol:      res.Symbol,
				ReservedQty: res.RemainingQty,
			}); err != nil {
				log.Error("release_shares_failed", zap.Error(err))
				return nil, status.Errorf(codes.Internal, "failed to release")
			}
		}
	}

	if err := qtx.DeleteReservation(ctx, orderUUID); err != nil {
		log.Error("delete_reservation_failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to release")
	}

	if err := tx.Commit(); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to release")
	}

	return &emptypb.Empty{}, nil
}
