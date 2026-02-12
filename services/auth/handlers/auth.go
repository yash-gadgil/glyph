package handlers

import (
	"context"
	"time"

	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/services/auth/utils"
	authpb "github.com/yash-gadgil/glyph/services/gen/golang/auth"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func invalidArg(err error, fallback string) error {
	if st, ok := status.FromError(err); ok {
		return status.Errorf(codes.InvalidArgument, "%s", st.Message())
	}
	return status.Error(codes.InvalidArgument, fallback)
}

func (s *AuthHandler) Signup(ctx context.Context, req *authpb.SignupRequest) (*emptypb.Empty, error) {
	log := logger.WithContextFields(ctx, s.log).With(logger.Action("signup"))

	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid signup request")
	}
	if err := utils.ValidateName(req.UserName); err != nil {
		return nil, invalidArg(err, "Invalid name")
	}
	if err := utils.ValidateEmail(req.Email); err != nil {
		return nil, invalidArg(err, "Invalid email")
	}
	if err := utils.ValidatePassword(req.Password); err != nil {
		return nil, invalidArg(err, "Invalid password")
	}

	log = log.With(logger.KV("user_email", req.Email))

	res, err := s.userClient.CheckEmailAvailability(ctx, &userpb.CheckEmailRequest{Email: req.Email})
	if err != nil {
		log.Error("email_availability_check_failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Signup failed")
	}
	if !res.Available {
		return nil, status.Errorf(codes.AlreadyExists, "Email in use")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("bcrypt_failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Signup failed")
	}
	if err := s.cache.StorePendingSignup(ctx, req.Email, req.UserName, string(hash), 30*time.Minute); err != nil {
		log.Error("store_pending_signup_failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Signup failed")
	}

	verifToken, err := utils.CreateTokenWithClaims(map[string]any{
		"email": req.Email,
	}, time.Now().Add(24*time.Hour), s.keyStore.GetCurrentKey())
	if err != nil {
		log.Error("verification_token_failed", zap.Error(err))
		_ = s.cache.DeletePendingSignup(ctx, req.Email)
		return nil, status.Errorf(codes.Internal, "Signup failed")
	}

	if err := s.cache.EnqueueVerificationEmail(ctx, req.Email, verifToken); err != nil {
		log.Error("enqueue_verification_email_failed", zap.Error(err))
		_ = s.cache.DeletePendingSignup(ctx, req.Email)
		return nil, status.Errorf(codes.Internal, "Signup failed")
	}

	log.Info("queued_verification_email")
	return &emptypb.Empty{}, nil
}

func (s *AuthHandler) Signin(ctx context.Context, req *authpb.SigninRequest) (*authpb.TokenResponse, error) {
	log := logger.WithContextFields(ctx, s.log).With(logger.Action("signin"))

	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid signin request")
	}
	if err := utils.ValidateEmail(req.Email); err != nil {
		return nil, invalidArg(err, "Invalid email")
	}
	if err := utils.ValidatePassword(req.Password); err != nil {
		return nil, invalidArg(err, "Invalid password")
	}

	log = log.With(logger.KV("user_email", req.Email))

	res, err := s.userClient.SigninUser(ctx, &userpb.SigninUserInfo{
		Email:    req.Email,
		Password: &req.Password,
	})
	if err != nil {
		log.Error("signin_failed", zap.Error(err))
		switch status.Code(err) {
		case codes.NotFound:
			return nil, status.Errorf(codes.NotFound, "We couldn't find an account for that email")
		case codes.PermissionDenied:
			return nil, status.Errorf(codes.PermissionDenied, "%s", status.Convert(err).Message())
		default:
			return nil, status.Errorf(codes.Unauthenticated, "Signin failed")
		}
	}

	return s.issueTokens(res.UserId, time.Hour, 30*24*time.Hour)
}

func (s *AuthHandler) issueTokens(userID string, accessTTL, refreshTTL time.Duration) (*authpb.TokenResponse, error) {
	accessToken, err := utils.CreateToken(userID, time.Now().Add(accessTTL), s.keyStore.GetCurrentKey())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create access token")
	}
	refreshToken, err := utils.CreateToken(userID, time.Now().Add(refreshTTL), s.keyStore.GetCurrentKey())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create refresh token")
	}
	return &authpb.TokenResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}
