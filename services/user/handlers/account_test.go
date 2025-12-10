package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
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

func TestSigninUserAcceptsValidPassword(t *testing.T) {
	h, mock := newAccountHandler(t)
	userID := uuid.New()
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	require.NoError(t, err)
	stored := string(hash)

	mock.ExpectQuery(`SELECT id, password_hash`).
		WithArgs("user@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "password_hash"}).AddRow(userID, stored))
	mock.ExpectExec(`INSERT INTO accounts`).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := h.SigninUser(context.Background(), &userpb.SigninUserInfo{
		Email:    "user@example.com",
		Password: strPtr("secret"),
	})
	require.NoError(t, err)
	assert.Equal(t, userID.String(), resp.UserId)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSigninUserRejectsWrongPassword(t *testing.T) {
	h, mock := newAccountHandler(t)
	userID := uuid.New()
	hash, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	stored := string(hash)

	mock.ExpectQuery(`SELECT id, password_hash`).
		WithArgs("user@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "password_hash"}).AddRow(userID, stored))

	_, err := h.SigninUser(context.Background(), &userpb.SigninUserInfo{
		Email:    "user@example.com",
		Password: strPtr("wrong"),
	})
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestSigninUserRejectsPasswordOnOAuthAccount(t *testing.T) {
	h, mock := newAccountHandler(t)
	userID := uuid.New()

	mock.ExpectQuery(`SELECT id, password_hash`).
		WithArgs("user@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "password_hash"}).AddRow(userID, nil))

	_, err := h.SigninUser(context.Background(), &userpb.SigninUserInfo{
		Email:    "user@example.com",
		Password: strPtr("secret"),
	})
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestSigninUserUnknownEmail(t *testing.T) {
	h, mock := newAccountHandler(t)

	mock.ExpectQuery(`SELECT id, password_hash`).
		WithArgs("missing@example.com").
		WillReturnError(sql.ErrNoRows)

	_, err := h.SigninUser(context.Background(), &userpb.SigninUserInfo{
		Email:    "missing@example.com",
		Password: strPtr("secret"),
	})
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestCheckEmailAvailabilityFreeEmail(t *testing.T) {
	h, mock := newAccountHandler(t)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("free@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	resp, err := h.CheckEmailAvailability(context.Background(), &userpb.CheckEmailRequest{
		Email: "free@example.com",
	})
	require.NoError(t, err)
	assert.True(t, resp.Available)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckEmailAvailabilityTakenEmail(t *testing.T) {
	h, mock := newAccountHandler(t)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("taken@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	resp, err := h.CheckEmailAvailability(context.Background(), &userpb.CheckEmailRequest{
		Email: "taken@example.com",
	})
	require.NoError(t, err)
	assert.False(t, resp.Available)
}

func TestUpdatePasswordByEmail(t *testing.T) {
	h, mock := newAccountHandler(t)

	mock.ExpectExec(`UPDATE users SET password_hash`).
		WithArgs("newhash", "user@example.com").
		WillReturnResult(sqlmock.NewResult(0, 1))

	_, err := h.UpdatePasswordByEmail(context.Background(), &userpb.UpdatePasswordRequest{
		Email:        "user@example.com",
		PasswordHash: "newhash",
	})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdatePasswordByEmailFails(t *testing.T) {
	h, mock := newAccountHandler(t)

	mock.ExpectExec(`UPDATE users SET password_hash`).
		WillReturnError(fmt.Errorf("db down"))

	_, err := h.UpdatePasswordByEmail(context.Background(), &userpb.UpdatePasswordRequest{
		Email:        "user@example.com",
		PasswordHash: "newhash",
	})
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestGetProfile(t *testing.T) {
	h, mock := newAccountHandler(t)
	userID := uuid.New()

	mock.ExpectQuery(`SELECT id, email, user_name FROM users`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "user_name"}).
			AddRow(userID, "user@example.com", "Yash"))

	resp, err := h.GetProfile(context.Background(), &userpb.UserSpecifier{UserId: userID.String()})
	require.NoError(t, err)
	assert.Equal(t, "user@example.com", resp.Email)
	assert.Equal(t, "Yash", resp.UserName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProfileInvalidID(t *testing.T) {
	h, _ := newAccountHandler(t)

	_, err := h.GetProfile(context.Background(), &userpb.UserSpecifier{UserId: "not-a-uuid"})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestGetProfileNotFound(t *testing.T) {
	h, mock := newAccountHandler(t)
	userID := uuid.New()

	mock.ExpectQuery(`SELECT id, email, user_name FROM users`).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	_, err := h.GetProfile(context.Background(), &userpb.UserSpecifier{UserId: userID.String()})
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestAddFundsIsRetired(t *testing.T) {
	h, _ := newAccountHandler(t)

	_, err := h.AddFunds(context.Background(), &userpb.AddFundsRequest{
		UserId:      uuid.New().String(),
		AmountCents: 5000,
	})
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestResetAccount(t *testing.T) {
	h, mock := newAccountHandler(t)
	userID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO accounts`).WithArgs(userID).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE accounts`).WithArgs(userID).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM positions`).WithArgs(userID).WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	_, err := h.ResetAccount(context.Background(), &userpb.UserSpecifier{UserId: userID.String()})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestResetAccountInvalidID(t *testing.T) {
	h, _ := newAccountHandler(t)

	_, err := h.ResetAccount(context.Background(), &userpb.UserSpecifier{UserId: "bad"})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestResetAccountRollsBackOnError(t *testing.T) {
	h, mock := newAccountHandler(t)
	userID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO accounts`).WithArgs(userID).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE accounts`).WithArgs(userID).WillReturnError(fmt.Errorf("db down"))
	mock.ExpectRollback()

	_, err := h.ResetAccount(context.Background(), &userpb.UserSpecifier{UserId: userID.String()})
	assert.Equal(t, codes.Internal, status.Code(err))
}
