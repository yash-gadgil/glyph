package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yash-gadgil/glyph/services/gateway/server/handlers"
	"github.com/yash-gadgil/glyph/services/gateway/tests/mocks"
	authpb "github.com/yash-gadgil/glyph/services/gen/golang/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func postJSON(path string, body any) *http.Request {
	buf, _ := json.Marshal(body)
	return httptest.NewRequest(http.MethodPost, path, bytes.NewReader(buf))
}

func cookieByName(resp *http.Response, name string) *http.Cookie {
	for _, c := range resp.Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func TestSignupReturnsOK(t *testing.T) {
	authClient := new(mocks.MockAuthClient)
	authClient.On("Signup", mock.Anything, mock.Anything).Return(&emptypb.Empty{}, nil)

	cfg := handlers.NewTestConfig(authClient)

	rec := httptest.NewRecorder()
	cfg.Signup(rec, postJSON("/auth/signup", map[string]string{
		"name": "Ada", "email": "ada@example.com", "password": "hunter2hunter2",
	}))

	assert.Equal(t, http.StatusOK, rec.Code)
	authClient.AssertExpectations(t)
}

func TestSigninSetsCookies(t *testing.T) {
	authClient := new(mocks.MockAuthClient)
	authClient.On("Signin", mock.Anything, mock.Anything).Return(&authpb.TokenResponse{
		AccessToken:  "access",
		RefreshToken: "refresh",
	}, nil)

	cfg := handlers.NewTestConfig(authClient)

	rec := httptest.NewRecorder()
	cfg.Signin(rec, postJSON("/auth/signin", map[string]string{
		"email": "ada@example.com", "password": "hunter2hunter2",
	}))

	resp := rec.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "access", cookieByName(resp, "accessToken").Value)
	assert.Equal(t, "refresh", cookieByName(resp, "refreshToken").Value)
}

func TestSigninNotFoundMapsTo404(t *testing.T) {
	authClient := new(mocks.MockAuthClient)
	authClient.On("Signin", mock.Anything, mock.Anything).Return(nil, status.Error(codes.NotFound, "no such user"))

	cfg := handlers.NewTestConfig(authClient)

	rec := httptest.NewRecorder()
	cfg.Signin(rec, postJSON("/auth/signin", map[string]string{
		"email": "missing@example.com", "password": "whatever",
	}))

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestSignoutClearsCookies(t *testing.T) {
	cfg := handlers.NewTestConfig(new(mocks.MockAuthClient))

	rec := httptest.NewRecorder()
	cfg.Signout(rec, httptest.NewRequest(http.MethodPost, "/auth/signout", nil))

	resp := rec.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, -1, cookieByName(resp, "accessToken").MaxAge)
	assert.Equal(t, -1, cookieByName(resp, "refreshToken").MaxAge)
}

func TestForgotPasswordAlwaysOK(t *testing.T) {
	authClient := new(mocks.MockAuthClient)
	authClient.On("ForgotPassword", mock.Anything, mock.Anything).Return(nil, status.Error(codes.NotFound, "no user"))

	cfg := handlers.NewTestConfig(authClient)

	rec := httptest.NewRecorder()
	cfg.ForgotPassword(rec, postJSON("/auth/forgot-password", map[string]string{"email": "x@y.z"}))

	assert.Equal(t, http.StatusOK, rec.Code)
}
