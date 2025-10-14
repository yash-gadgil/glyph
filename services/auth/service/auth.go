package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yash-gadgil/glyph/services/auth/types"
	"github.com/yash-gadgil/glyph/services/auth/utils"
	authpb "github.com/yash-gadgil/glyph/services/gen/golang/auth"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"golang.org/x/oauth2"
)

type AuthService struct {
	googleConfig *oauth2.Config
	AddrConfig   *types.AddrConfig
}

func NewAuthService(cfg *oauth2.Config, acfg *types.AddrConfig) *AuthService {
	return &AuthService{
		googleConfig: cfg,
		AddrConfig:   acfg,
	}
}

func (s *AuthService) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	fmt.Println("Registering User:", req.Email, "with password", req.Password)
	return &authpb.RegisterResponse{}, nil
}
func (s *AuthService) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	return &authpb.LoginResponse{}, nil
}
func (s *AuthService) OAuthURL(ctx context.Context, req *authpb.OAuthURLRequest) (*authpb.OAuthURLResponse, error) {

	url := s.googleConfig.AuthCodeURL(req.Status, oauth2.AccessTypeOffline)

	fmt.Println("URL Requested")

	return &authpb.OAuthURLResponse{
		Url: url,
	}, nil
}

func (s *AuthService) OAuthCallback(ctx context.Context, req *authpb.OAuthCallbackRequest) (*authpb.OAuthCallbackResponse, error) {

	fmt.Println("Callback")

	t, err := s.googleConfig.Exchange(ctx, req.Code)
	if err != nil {
		return &authpb.OAuthCallbackResponse{}, err
	}

	client := s.googleConfig.Client(ctx, t)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return &authpb.OAuthCallbackResponse{}, err
	}

	var jsonResp types.GoAuth
	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		return &authpb.OAuthCallbackResponse{}, err
	}

	fmt.Println(jsonResp.Email)

	serverAddr := s.AddrConfig.UserSvcAddr
	conn := utils.GetGrpcClient(serverAddr)
	defer conn.Close()

	c := userpb.NewAccountServiceClient(conn)
	ctx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	res, err := c.RegisterUser(ctx, &userpb.UserInfo{
		Email: jsonResp.Email,
	})
	if err != nil {
		fmt.Println("Error in Server Response:", err)
	}
	fmt.Println(res)

	return &authpb.OAuthCallbackResponse{
		Token: t.AccessToken,
	}, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, req *authpb.EmailVerificationRequest) (*authpb.EmailVerificationResponse, error) {
	return &authpb.EmailVerificationResponse{}, nil
}
