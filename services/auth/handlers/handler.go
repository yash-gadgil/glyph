package handlers

import (
	"context"
	"fmt"

	"github.com/yash-gadgil/glyph/services/auth/types"
	authpb "github.com/yash-gadgil/glyph/services/gen/golang/auth"
	"google.golang.org/grpc"
)

type AuthHandler struct {
	authService types.AuthService
	authpb.UnimplementedAuthServiceServer
}

func NewGrpcAuthService(grpc *grpc.Server, authService types.AuthService) {

	handler := &AuthHandler{
		authService: authService,
	}
	authpb.RegisterAuthServiceServer(grpc, handler)
}

func (h *AuthHandler) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	fmt.Println("Registering User:", req.Email, "with password", req.Password)
	return &authpb.RegisterResponse{}, nil
}
func (h *AuthHandler) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	return &authpb.LoginResponse{}, nil
}
func (h *AuthHandler) OAuthURL(ctx context.Context, req *authpb.OAuthURLRequest) (*authpb.OAuthURLResponse, error) {
	res, err := h.authService.OAuthURL(ctx, req)
	if err != nil {
		return &authpb.OAuthURLResponse{}, err
	}
	return res, nil
}

func (h *AuthHandler) OAuthCallback(ctx context.Context, req *authpb.OAuthCallbackRequest) (*authpb.OAuthCallbackResponse, error) {
	res, err := h.authService.OAuthCallback(ctx, req)
	if err != nil {
		return &authpb.OAuthCallbackResponse{}, err
	}
	return res, nil
}

func (h *AuthHandler) VerifyEmail(ctx context.Context, req *authpb.EmailVerificationRequest) (*authpb.EmailVerificationResponse, error) {
	return &authpb.EmailVerificationResponse{}, nil
}
