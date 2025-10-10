all: proto

proto_u:
	protoc -I proto \
		--go_out=paths=source_relative:./services/gen/golang/user \
		--go-grpc_out=paths=source_relative:./services/gen/golang/user \
		proto/user.proto
#go -C ./services/user mod tidy

genUser:
	@protoc -I proto \
	--proto_path=proto "proto/user.proto" \
	--go_out=services/gen/golang/user --go_opt=paths=source_relative \
	--go-grpc_out=services/gen/golang/user --go-grpc_opt=paths=source_relative