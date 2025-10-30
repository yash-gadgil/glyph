package service

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/yash-gadgil/glyph/services/auth/db"
	"github.com/yash-gadgil/glyph/services/auth/types"
	"github.com/yash-gadgil/glyph/services/auth/utils"
	authpb "github.com/yash-gadgil/glyph/services/gen/golang/auth"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService struct {
	googleConfig *oauth2.Config
	AddrConfig   *types.AddrConfig

	userConn   *grpc.ClientConn
	userClient userpb.AccountServiceClient

	cache *db.Cache
}

func NewAuthService(cfg *oauth2.Config, acfg *types.AddrConfig, cache *db.Cache) *AuthService {
	svc := &AuthService{
		googleConfig: cfg,
		AddrConfig:   acfg,
		cache:        cache,
	}
	if acfg != nil && acfg.UserSvcAddr != "" {
		conn := utils.GetGrpcClient(acfg.UserSvcAddr)
		if conn != nil {
			svc.userConn = conn
			svc.userClient = userpb.NewAccountServiceClient(conn)
		}
	}

	return svc
}

func (s *AuthService) Close() error {
	if s.userConn != nil {
		return s.userConn.Close()
	}
	return nil
}

func (s *AuthService) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {

	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	res, err := s.userClient.CheckEmailAvailability(ctx, &userpb.CheckEmailRequest{
		Email: req.Email,
	})
	if err != nil {
		log.Printf("email availability check failed: %v", err) // LOG
		return nil, status.Errorf(codes.Internal, "Failed to process registration")
	}
	if !res.Available {
		return nil, status.Errorf(codes.AlreadyExists, "Email already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("bcrypt failed: %v", err) // LOG
		return nil, status.Errorf(codes.Internal, "Failed to process registration")
	}
	if err := s.cache.StorePendingRegistration(ctx, req.Email, string(hash), 30*time.Minute); err != nil {
		log.Printf("failed to store pending registration: %v", err) // LOG
		return nil, status.Errorf(codes.Internal, "Failed to process registration")
	}

	go func(email string) {
		if err := utils.SendEmail(email, "Verify your email", "Click to verify your account"); err != nil {
			log.Printf("failed to send verification email to %s: %v", email, err) // LOG
		} else {
			log.Printf("verification email sent to %s", email) // LOG
		}
	}(req.Email)

	return &authpb.RegisterResponse{}, nil
}

func (s *AuthService) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {

	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	res, err := s.userClient.LoginUser(ctx, &userpb.UserInfo{
		Email:    req.Email,
		Password: &req.Password,
	})
	if err != nil {
		log.Printf("incorrect pass or email: %v", err) // LOG
		return nil, status.Errorf(codes.Unauthenticated, "Failed to process login")
	}

	accessToken, err := utils.CreateToken(res.UserID, time.Now().Add(time.Hour))
	if err != nil {
		return &authpb.LoginResponse{}, err
	}

	log.Println("Logging In")

	refreshToken, err := utils.CreateToken(res.UserID, time.Now().Add(time.Hour*24))
	if err != nil {
		return &authpb.LoginResponse{}, err
	}

	return &authpb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) OAuthURL(ctx context.Context, req *authpb.OAuthURLRequest) (*authpb.OAuthURLResponse, error) {

	url := s.googleConfig.AuthCodeURL(req.State, oauth2.AccessTypeOffline)

	log.Println("URL Requested") // LOG

	return &authpb.OAuthURLResponse{
		Url: url,
	}, nil
}

func (s *AuthService) OAuthCallback(ctx context.Context, req *authpb.OAuthCallbackRequest) (*authpb.OAuthCallbackResponse, error) {

	log.Println("Callback") // LOG TO REMOVE

	t, err := s.googleConfig.Exchange(ctx, req.Code)
	if err != nil {
		return nil, err
	}

	client := s.googleConfig.Client(ctx, t)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}

	var jsonResp types.GoAuth
	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		return nil, err
	}

	log.Println(jsonResp.Email) // LOG TO REMOVE

	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	var userID string

	switch req.State {
	case "register":
		res, err := s.userClient.RegisterUser(ctx, &userpb.UserInfo{
			Email:    jsonResp.Email,
			Password: nil,
		})
		if err != nil {
			log.Println("Error in Server Response:", err) // LOG
			return nil, status.Errorf(codes.Internal, "Failed to process register")
		}
		userID = res.GetUserID()
	case "login":
		res, err := s.userClient.LoginUser(ctx, &userpb.UserInfo{
			Email:    jsonResp.Email,
			Password: nil,
		})
		if err != nil {
			log.Println("Error in Server Response:", err) // LOG
			return nil, status.Errorf(codes.Unauthenticated, "Failed to process login")
		}
		userID = res.GetUserID()
	default:
		return nil, status.Errorf(codes.InvalidArgument, "Invalid state passed in callback: %s", req.State)
	}

	accessToken, err := utils.CreateToken(userID, time.Now().Add(time.Hour))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create access token")
	}

	refreshToken, err := utils.CreateToken(userID, time.Now().Add(24*time.Hour))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create refresh token")
	}

	return &authpb.OAuthCallbackResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, req *authpb.EmailVerificationRequest) (*authpb.EmailVerificationResponse, error) {

	claims, err := utils.ParseTokenClaims(req.Token)
	if err != nil {
		log.Printf("invalid verification token: %v", err) // LOG
		return nil, status.Errorf(codes.InvalidArgument, "invalid or expired verification link")
	}

	email, ok := claims["email"].(string)
	if !ok || email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email claim missing or invalid")
	}

	passwordHash, err := s.cache.GetPendingRegistration(ctx, email)
	if err != nil {
		log.Printf("failed to get pending registration for %s: %v", email, err) // LOG
		return nil, status.Errorf(codes.NotFound, "verification link invalid or expired")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := s.userClient.RegisterUser(ctx, &userpb.UserInfo{
		Email:    email,
		Password: &passwordHash,
	})
	if err != nil {
		log.Printf("failed to create user: %v", err) // LOG
		return nil, status.Errorf(codes.Internal, "failed to complete registration")
	}

	if err := s.cache.DeletePendingRegistration(ctx, email); err != nil {
		log.Printf("failed to delete pending registration for %s: %v", email, err) // LOG
	}

	userID := res.GetUserID()
	if userID == "" {
		return nil, status.Errorf(codes.Internal, "user ID not available")
	}

	accessToken, err := utils.CreateToken(userID, time.Now().Add(time.Hour*3))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create access token")
	}

	refreshToken, err := utils.CreateToken(userID, time.Now().Add(24*time.Hour*7))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create refresh token")
	}

	return &authpb.EmailVerificationResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) VerifyToken(ctx context.Context, req *authpb.VerificationRequest) (*authpb.VerificationResponse, error) {
	userId, err := utils.VerifyToken(req.Token)
	if err != nil {
		return nil, err
	}
	return &authpb.VerificationResponse{
		UserID: userId,
	}, nil
}
