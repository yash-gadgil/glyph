all: proto

genUser:
	@protoc -I proto \
	--proto_path=proto "proto/user.proto" \
	--go_out=services/gen/golang/user --go_opt=paths=source_relative \
	--go-grpc_out=services/gen/golang/user --go-grpc_opt=paths=source_relative

genAuth:
	@protoc -I proto \
	--proto_path=proto "proto/auth.proto" \
	--go_out=services/gen/golang/auth --go_opt=paths=source_relative \
	--go-grpc_out=services/gen/golang/auth --go-grpc_opt=paths=source_relative

genMrkt:
	@protoc -I proto \
	--proto_path=proto "proto/mrktdata.proto" \
	--go_out=services/gen/golang/mrktdata --go_opt=paths=source_relative \
	--go-grpc_out=services/gen/golang/mrktdata --go-grpc_opt=paths=source_relative