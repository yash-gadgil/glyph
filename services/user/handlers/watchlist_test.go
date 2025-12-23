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

func TestCreateWatchlist(t *testing.T) {
	h, mock := newWatchlistHandler(t)
	userID := uuid.New()
	listID := uuid.New()
	name := "Tech"

	mock.ExpectQuery(`INSERT INTO watchlists`).
		WithArgs(name, userID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(listID))

	resp, err := h.CreateWatchlist(context.Background(), &userpb.CreateWatchlistRequest{
		UserId: userID.String(),
		Name:   &name,
	})
	require.NoError(t, err)
	assert.Equal(t, listID.String(), resp.Id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateWatchlistDuplicate(t *testing.T) {
	h, mock := newWatchlistHandler(t)
	userID := uuid.New()
	name := "Tech"

	mock.ExpectQuery(`INSERT INTO watchlists`).
		WillReturnError(fmt.Errorf("pq: duplicate key value violates unique constraint"))

	_, err := h.CreateWatchlist(context.Background(), &userpb.CreateWatchlistRequest{
		UserId: userID.String(),
		Name:   &name,
	})
	assert.Equal(t, codes.AlreadyExists, status.Code(err))
}

func TestCreateWatchlistEmptyName(t *testing.T) {
	h, _ := newWatchlistHandler(t)
	name := "   "

	_, err := h.CreateWatchlist(context.Background(), &userpb.CreateWatchlistRequest{
		UserId: uuid.New().String(),
		Name:   &name,
	})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestModifyWatchlistSubscribe(t *testing.T) {
	h, mock := newWatchlistHandler(t)
	userID := uuid.New()
	listID := uuid.New()

	mock.ExpectQuery(`INSERT INTO watchlist_symbols`).
		WillReturnRows(sqlmock.NewRows([]string{"watchlist_id", "symbol"}).AddRow(listID, "AAPL"))

	_, err := h.ModifyWatchlist(context.Background(), &userpb.ModifyWatchlistRequest{
		Id:      listID.String(),
		UserId:  userID.String(),
		Action:  userpb.ModifyWatchlistRequest_SUBSCRIBE,
		Symbols: []string{"aapl"},
	})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestModifyWatchlistUnsubscribe(t *testing.T) {
	h, mock := newWatchlistHandler(t)
	userID := uuid.New()
	listID := uuid.New()

	mock.ExpectQuery(`DELETE FROM watchlist_symbols`).
		WillReturnRows(sqlmock.NewRows([]string{"watchlist_id", "symbol"}).AddRow(listID, "AAPL"))

	_, err := h.ModifyWatchlist(context.Background(), &userpb.ModifyWatchlistRequest{
		Id:      listID.String(),
		UserId:  userID.String(),
		Action:  userpb.ModifyWatchlistRequest_UNSUBSCRIBE,
		Symbols: []string{"AAPL"},
	})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestModifyWatchlistNoSymbols(t *testing.T) {
	h, _ := newWatchlistHandler(t)

	_, err := h.ModifyWatchlist(context.Background(), &userpb.ModifyWatchlistRequest{
		Id:      uuid.New().String(),
		UserId:  uuid.New().String(),
		Action:  userpb.ModifyWatchlistRequest_SUBSCRIBE,
		Symbols: []string{},
	})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestDeleteWatchlist(t *testing.T) {
	h, mock := newWatchlistHandler(t)
	userID := uuid.New()
	listID := uuid.New()

	mock.ExpectExec(`DELETE FROM watchlists`).
		WithArgs(listID, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	_, err := h.DeleteWatchlist(context.Background(), &userpb.WatchlistSpecifier{
		Id:     listID.String(),
		UserId: userID.String(),
	})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteWatchlistInvalidID(t *testing.T) {
	h, _ := newWatchlistHandler(t)

	_, err := h.DeleteWatchlist(context.Background(), &userpb.WatchlistSpecifier{
		Id:     "bad",
		UserId: uuid.New().String(),
	})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}
