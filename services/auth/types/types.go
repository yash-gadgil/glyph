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
}

type GoAuth struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

type AddrConfig struct {
	UserSvcAddr string
}
