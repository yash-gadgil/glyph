package main

import (
	"os"

	"github.com/yash-gadgil/glyph/services/auth/server"
)

func main() {

	grpcServer := server.NewGrpcServer(os.Getenv("AUTH_SVC_PORT"))
	grpcServer.Run()
}
