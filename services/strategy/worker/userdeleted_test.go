package worker

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gen "github.com/yash-gadgil/glyph/services/strategy/db/gen"
	"go.uber.org/zap"
)

func newQueries(t *testing.T) (*gen.Queries, sqlmock.Sqlmock) {
	t.Helper()
	sdb, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = sdb.Close() })
	return gen.New(sdb), mock
}

func TestHandleUserDeletedDeletesStrategies(t *testing.T) {
	q, mock := newQueries(t)
	userID := uuid.New()

	mock.ExpectExec(`DELETE FROM strategies`).WithArgs(userID).WillReturnResult(sqlmock.NewResult(0, 3))

	body := []byte(`{"user_id":"` + userID.String() + `"}`)
	err := handleUserDeleted(context.Background(), q, zap.NewNop(), body)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleUserDeletedBadPayloadIsPermanent(t *testing.T) {
	q, _ := newQueries(t)

	err := handleUserDeleted(context.Background(), q, zap.NewNop(), []byte(`not json`))
	assert.True(t, errors.As(err, &permanentError{}))
}

func TestHandleUserDeletedBadUserIDIsPermanent(t *testing.T) {
	q, _ := newQueries(t)

	err := handleUserDeleted(context.Background(), q, zap.NewNop(), []byte(`{"user_id":"nope"}`))
	assert.True(t, errors.As(err, &permanentError{}))
}
