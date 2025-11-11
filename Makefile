.DEFAULT_GOAL := build

proto:
	buf generate

build:
	go build ./...
	cargo build

test:
	go test ./...
	cargo test

fmt:
	gofmt -w .
	cargo fmt

vet:
	go vet ./...

lint: vet
	cargo clippy

.PHONY: proto build test fmt vet lint
