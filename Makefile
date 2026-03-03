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

web-install:
	cd web && pnpm install

web-build:
	cd web && pnpm build

web-test:
	cd web && pnpm test

web-lint:
	cd web && pnpm lint

.PHONY: proto build test fmt vet lint web-install web-build web-test web-lint
