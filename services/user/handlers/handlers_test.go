package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"google.golang.org/grpc"
)

func TestUserHandlersRegisterOnGrpcServer(t *testing.T) {
	srv := grpc.NewServer()
	userpb.RegisterAccountServiceServer(srv, NewAccountHandler(nil, nil))
	userpb.RegisterPortfolioServiceServer(srv, NewPortfolioHandler(nil, nil, nil))
	userpb.RegisterWatchlistServiceServer(srv, NewWatchlistHandler(nil, nil))

	info := srv.GetServiceInfo()
	for _, svc := range []string{"user.AccountService", "user.PortfolioService", "user.WatchlistService"} {
		_, ok := info[svc]
		assert.True(t, ok, "%s should be registered", svc)
	}
}
