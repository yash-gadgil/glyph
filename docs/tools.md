# Tools

A few command line tools are not Go or Rust dependencies, so you install them once. They
handle proto codegen, the database layer, and migrations.

## Proto codegen: buf and the Go plugins

The Go and Python gRPC stubs are generated with buf. buf reads `buf.gen.yaml` at the repo
root, finds every `.proto` under `proto/`, and writes the Go stubs into
`services/gen/golang/` and the Python stubs (for the inference service) into
`services/gen/python/`.

```
brew install bufbuild/buf/buf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Regenerate everything after editing a proto:

```
make proto
```

The Rust orderbook does not use buf. It generates its own server stubs at build time
through `build.rs` and `tonic-build`, so a plain `cargo build` is enough there.

## Database: goose and sqlc

goose runs the SQL migrations in `services/<service>/db/migrations`. sqlc generates the Go
query layer from each service's `schema.sql` and its `queries` folder.

```
go install github.com/pressly/goose/v3/cmd/goose@v3.26.0
go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0
```

Both land in `$(go env GOPATH)/bin`, so make sure that is on your `PATH`.

### Regenerate the query code

sqlc reads `schema.sql`, not the migrations, so when you change a query or a table you
edit both `queries/*.sql` and `schema.sql`, then regenerate:

```
cd services/user/db
sqlc generate
```

The generated code under `db/gen/` is committed.

### Run migrations

Point goose at a service's migrations folder and your Postgres connection:

```
goose -dir services/user/db/migrations postgres "postgres://USER:PASS@localhost:5432/userdb?sslmode=disable" up
```

In the Kubernetes stack the migrations run automatically through an init container on
deploy, so you only run goose by hand for local development.

## Rust

The orderbook needs a stable Rust toolchain with clippy and rustfmt:

```
rustup toolchain install stable
rustup component add clippy rustfmt
```

## Python: uv

The inference service manages its dependencies with [uv](https://github.com/astral-sh/uv)
instead of pip, including a CPU-only build of torch on Linux so the built image stays
small.

```
curl -LsSf https://astral.sh/uv/install.sh | sh
cd services/inference && uv sync --extra model
```

The `model` extra pulls in `transformers` and `torch`; without it the service still runs,
just permanently in its echo fallback mode.
