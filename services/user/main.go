package main

import "github.com/yash-gadgil/glyph/services/user/server"

func main() {

	grpcServer := server.NewGRPCServer(":9000")
	grpcServer.Run()

}
