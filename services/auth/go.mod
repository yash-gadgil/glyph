module github.com/yash-gadgil/glyph/services/auth

go 1.25.1

require (
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/redis/go-redis/v9 v9.14.1
	github.com/yash-gadgil/glyph/services/gen/golang v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.43.0
	golang.org/x/oauth2 v0.30.0
	google.golang.org/grpc v1.76.0
)

require (
	cloud.google.com/go/compute/metadata v0.7.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	golang.org/x/net v0.45.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace github.com/yash-gadgil/glyph/services/gen/golang => ../gen/golang
