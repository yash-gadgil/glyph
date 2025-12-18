package handlers

import (
	"context"
	"database/sql"
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

func newWatchlistHandler(t *testing.T) (*WatchlistHandler, sqlmock.Sqlmock) {
	t.Helper()
	sdb, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = sdb.Close() })
	return NewWatchlistHandler(sdb, zap.NewNop()), mock
}

func TestGetWatchlistsReturnsMetadataAndFirst(t *testing.T) {
	h, mock := newWatchlistHandler(t)
	userID := uuid.New()
	listID := uuid.New()

	mock.ExpectQuery(`SELECT id, w_name`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "w_name"}).AddRow(listID, "Tech"))
	mock.ExpectQuery(`FROM watchlists w`).
		WithArgs(listID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "w_name", "symbols"}).
			AddRow(userID, "Tech", []byte("{AAPL,MSFT}")))

	resp, err := h.GetWatchlists(context.Background(), &userpb.UserSpecifier{UserId: userID.String()})
	require.NoError(t, err)
	require.Len(t, resp.WMetadata, 1)
	assert.Equal(t, "Tech", resp.WMetadata[0].Name)
	require.NotNil(t, resp.First)
	assert.Equal(t, []string{"AAPL", "MSFT"}, resp.First.Symbols)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetWatchlistsEmpty(t *testing.T) {
	h, mock := newWatchlistHandler(t)
	userID := uuid.New()

	mock.ExpectQuery(`SELECT id, w_name`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "w_name"}))

	resp, err := h.GetWatchlists(context.Background(), &userpb.UserSpecifier{UserId: userID.String()})
	require.NoError(t, err)
	assert.Empty(t, resp.WMetadata)
	assert.Nil(t, resp.First)
}

func TestGetWatchlistsInvalidUser(t *testing.T) {
	h, _ := newWatchlistHandler(t)

	_, err := h.GetWatchlists(context.Background(), &userpb.UserSpecifier{UserId: "bad"})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestGetWatchlistReturnsSymbols(t *testing.T) {
	h, mock := newWatchlistHandler(t)
	userID := uuid.New()
	listID := uuid.New()

	mock.ExpectQuery(`FROM watchlists w`).
		WithArgs(listID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "w_name", "symbols"}).
			AddRow(userID, "Tech", []byte("{AAPL}")))

	resp, err := h.GetWatchlist(context.Background(), &userpb.WatchlistSpecifier{
		Id:     listID.String(),
		UserId: userID.String(),
	})
	require.NoError(t, err)
	assert.Equal(t, "Tech", resp.Name)
	assert.Equal(t, []string{"AAPL"}, resp.Symbols)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetWatchlistNotFound(t *testing.T) {
	h, mock := newWatchlistHandler(t)
	userID := uuid.New()
	listID := uuid.New()

	mock.ExpectQuery(`FROM watchlists w`).
		WithArgs(listID, userID).
		WillReturnError(sql.ErrNoRows)

	_, err := h.GetWatchlist(context.Background(), &userpb.WatchlistSpecifier{
		Id:     listID.String(),
		UserId: userID.String(),
	})
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestGetWatchlistInvalidID(t *testing.T) {
	h, _ := newWatchlistHandler(t)

	_, err := h.GetWatchlist(context.Background(), &userpb.WatchlistSpecifier{
		Id:     "bad",
		UserId: uuid.New().String(),
	})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}
