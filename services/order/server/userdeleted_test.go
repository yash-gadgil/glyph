package server

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestHandleUserDeletedDeletesOrdersAndFills(t *testing.T) {
	sdb, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer sdb.Close()
	userID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM fills`).WithArgs(userID).WillReturnResult(sqlmock.NewResult(0, 5))
	mock.ExpectExec(`DELETE FROM orders`).WithArgs(userID).WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	err = handleUserDeleted(context.Background(), sdb, zap.NewNop(), []byte(`{"user_id":"`+userID.String()+`"}`))
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleUserDeletedRollsBackOnError(t *testing.T) {
	sdb, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer sdb.Close()
	userID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM fills`).WithArgs(userID).WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	err = handleUserDeleted(context.Background(), sdb, zap.NewNop(), []byte(`{"user_id":"`+userID.String()+`"}`))
	require.Error(t, err)
	assert.False(t, errors.As(err, &userDeletedPermanentError{}))
}

func TestHandleUserDeletedBadPayloadIsPermanent(t *testing.T) {
	sdb, _, err := sqlmock.New()
	require.NoError(t, err)
	defer sdb.Close()

	err = handleUserDeleted(context.Background(), sdb, zap.NewNop(), []byte(`not json`))
	assert.True(t, errors.As(err, &userDeletedPermanentError{}))
}
