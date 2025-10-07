.DEFAULT_GOAL := build

proto:
	buf generate

build:
	go build ./...

test:
	go test ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

lint: vet

.PHONY: proto build test fmt vet lint
