package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yash-gadgil/glyph/services/auth/db"
	"github.com/yash-gadgil/glyph/services/auth/utils"
	authpb "github.com/yash-gadgil/glyph/services/gen/golang/auth"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type mockAccountClient struct {
	mock.Mock
	userpb.AccountServiceClient
}

func (m *mockAccountClient) CheckEmailAvailability(ctx context.Context, in *userpb.CheckEmailRequest, opts ...grpc.CallOption) (*userpb.CheckEmailResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*userpb.CheckEmailResponse), args.Error(1)
}

func (m *mockAccountClient) SigninUser(ctx context.Context, in *userpb.SigninUserInfo, opts ...grpc.CallOption) (*userpb.UserSpecifier, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*userpb.UserSpecifier), args.Error(1)
}

func (m *mockAccountClient) SignupUser(ctx context.Context, in *userpb.SignupUserInfo, opts ...grpc.CallOption) (*userpb.UserSpecifier, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*userpb.UserSpecifier), args.Error(1)
}

func (m *mockAccountClient) UpdatePasswordByEmail(ctx context.Context, in *userpb.UpdatePasswordRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func newAuthHandler(t *testing.T) (*AuthHandler, *mockAccountClient, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cache := &db.Cache{Rdb: rdb}
	ks, err := utils.NewKeyStore(zap.NewNop())
	require.NoError(t, err)
	client := new(mockAccountClient)
	return NewTestAuthHandler(nil, cache, ks, client, zap.NewNop()), client, mr
}

func TestSignupQueuesVerification(t *testing.T) {
	h, client, mr := newAuthHandler(t)

	client.On("CheckEmailAvailability", mock.Anything, mock.Anything).
		Return(&userpb.CheckEmailResponse{Available: true}, nil)

	_, err := h.Signup(context.Background(), &authpb.SignupRequest{
		UserName: "Yash", Email: "user@example.com", Password: "Passw0rd!",
	})
	require.NoError(t, err)

	entries, err := mr.List("email_verification_queue")
	require.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Contains(t, entries[0], "user@example.com")
}

func TestSignupEmailInUse(t *testing.T) {
	h, client, _ := newAuthHandler(t)

	client.On("CheckEmailAvailability", mock.Anything, mock.Anything).
		Return(&userpb.CheckEmailResponse{Available: false}, nil)

	_, err := h.Signup(context.Background(), &authpb.SignupRequest{
		UserName: "Yash", Email: "user@example.com", Password: "Passw0rd!",
	})
	assert.Equal(t, codes.AlreadyExists, status.Code(err))
}

func TestSignupRejectsWeakPassword(t *testing.T) {
	h, _, _ := newAuthHandler(t)

	_, err := h.Signup(context.Background(), &authpb.SignupRequest{
		UserName: "Yash", Email: "user@example.com", Password: "weak",
	})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestSigninIssuesTokens(t *testing.T) {
	h, client, _ := newAuthHandler(t)

	client.On("SigninUser", mock.Anything, mock.Anything).
		Return(&userpb.UserSpecifier{UserId: "user-1"}, nil)

	resp, err := h.Signin(context.Background(), &authpb.SigninRequest{
		Email: "user@example.com", Password: "Passw0rd!",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)

	uid, err := utils.VerifyToken(resp.AccessToken, h.keyStore)
	require.NoError(t, err)
	assert.Equal(t, "user-1", uid)
}

func TestSigninUnknownEmail(t *testing.T) {
	h, client, _ := newAuthHandler(t)

	client.On("SigninUser", mock.Anything, mock.Anything).
		Return((*userpb.UserSpecifier)(nil), status.Error(codes.NotFound, "no user"))

	_, err := h.Signin(context.Background(), &authpb.SigninRequest{
		Email: "user@example.com", Password: "Passw0rd!",
	})
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestVerifyEmailPromotesPending(t *testing.T) {
	h, client, _ := newAuthHandler(t)
	ctx := context.Background()

	require.NoError(t, h.cache.StorePendingSignup(ctx, "user@example.com", "Yash", "hash", 30*time.Minute))
	token, err := utils.CreateTokenWithClaims(map[string]any{"email": "user@example.com"},
		time.Now().Add(time.Hour), h.keyStore.GetCurrentKey())
	require.NoError(t, err)

	client.On("SignupUser", mock.Anything, mock.Anything).
		Return(&userpb.UserSpecifier{UserId: "user-1"}, nil)

	resp, err := h.VerifyEmail(ctx, &authpb.VerificationRequest{Token: token})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
}

func TestVerifyEmailBadToken(t *testing.T) {
	h, _, _ := newAuthHandler(t)
	_, err := h.VerifyEmail(context.Background(), &authpb.VerificationRequest{Token: "garbage"})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}
