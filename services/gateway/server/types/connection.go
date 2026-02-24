package types

import (
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type Connection[T any] struct {
	Conn   *grpc.ClientConn
	Client T
}

func NewConnection[T any](addr string, clientFactory func(*grpc.ClientConn) T) *Connection[T] {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Second * 10,
			Timeout:             time.Second * 3,
			PermitWithoutStream: true,
		}),
	}

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		log.Fatalf("failed to connect to %s: %v", addr, err)
	}

	return &Connection[T]{
		Conn:   conn,
		Client: clientFactory(conn),
	}
}

func (c *Connection[T]) Close() error {
	if c.Conn != nil {
		return c.Conn.Close()
	}
	return nil
}
