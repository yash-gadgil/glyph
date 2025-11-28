package handlers

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newAccountHandler(t *testing.T) (*AccountHandler, sqlmock.Sqlmock) {
	t.Helper()
	sdb, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = sdb.Close() })
	return NewAccountHandler(sdb, zap.NewNop()), mock
}

func strPtr(s string) *string { return &s }

func TestSignupUserCreatesUserAndAccount(t *testing.T) {
	h, mock := newAccountHandler(t)
	userID := uuid.New()
	accountID := uuid.New()

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("user@example.com", strPtr("hashed"), "Yash").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))
	mock.ExpectQuery(`INSERT into accounts`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(accountID))

	resp, err := h.SignupUser(context.Background(), &userpb.SignupUserInfo{
		UserName: "Yash",
		Email:    "user@example.com",
		Password: strPtr("hashed"),
	})
	require.NoError(t, err)
	assert.Equal(t, userID.String(), resp.UserId)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSignupUserFailsWhenUserInsertFails(t *testing.T) {
	h, mock := newAccountHandler(t)

	mock.ExpectQuery(`INSERT INTO users`).
		WillReturnError(fmt.Errorf("unique constraint violation"))

	_, err := h.SignupUser(context.Background(), &userpb.SignupUserInfo{
		UserName: "Yash",
		Email:    "user@example.com",
		Password: strPtr("hashed"),
	})
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestSignupUserFailsWhenAccountInsertFails(t *testing.T) {
	h, mock := newAccountHandler(t)
	userID := uuid.New()

	mock.ExpectQuery(`INSERT INTO users`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))
	mock.ExpectQuery(`INSERT into accounts`).
		WillReturnError(fmt.Errorf("db down"))

	_, err := h.SignupUser(context.Background(), &userpb.SignupUserInfo{
		UserName: "Yash",
		Email:    "user@example.com",
		Password: strPtr("hashed"),
	})
	assert.Equal(t, codes.Internal, status.Code(err))
}
