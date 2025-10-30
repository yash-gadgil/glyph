package types

import (
	"context"

	authpb "github.com/yash-gadgil/glyph/services/gen/golang/auth"
)

type AuthService interface {
	Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error)

	Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error)

	OAuthURL(ctx context.Context, req *authpb.OAuthURLRequest) (*authpb.OAuthURLResponse, error)

	OAuthCallback(ctx context.Context, req *authpb.OAuthCallbackRequest) (*authpb.OAuthCallbackResponse, error)

	VerifyEmail(ctx context.Context, req *authpb.EmailVerificationRequest) (*authpb.EmailVerificationResponse, error)

	VerifyToken(ctx context.Context, req *authpb.VerificationRequest) (*authpb.VerificationResponse, error)
}

type GoAuth struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
}

type AddrConfig struct {
	UserSvcAddr string
}
