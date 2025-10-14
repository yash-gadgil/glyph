module github.com/yash-gadgil/glyph/services/auth

go 1.25.1

require (
	github.com/joho/godotenv v1.5.1
	github.com/yash-gadgil/glyph/services/gen/golang v0.0.0-00010101000000-000000000000
	golang.org/x/oauth2 v0.30.0
	google.golang.org/grpc v1.76.0
)

require (
	cloud.google.com/go/compute/metadata v0.7.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace github.com/yash-gadgil/glyph/services/gen/golang => ../gen/golang
