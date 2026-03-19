# Setup

This gets the whole stack running locally. There are two ways to run it: Tilt for the
full Kubernetes stack, or by hand if you only need a few pieces.

## Prerequisites

- Go (see the version in `go.mod`)
- Rust (stable) for the orderbook
- Node and pnpm for the web app
- Docker
- The extra build tools in [tools.md](tools.md): goose, sqlc, buf, and the protoc Go plugins
- An Alpaca account for market data keys (the paper trading keys are free)

## Environment

Copy the example env file and fill it in:

```
cp .env.example .env
```

You need, at a minimum:

- `APCA_API_KEY_ID` and `APCA_API_SECRET_KEY` for Alpaca market data
- the service ports (the defaults in `.env.example` are fine)
- the Postgres connection settings for `userdb` and `orderdb`
- `RABBITMQ_URL` if you want fills to settle
- the auth secrets (JWT keys, Google OAuth, SMTP) only if you are working on auth

If you skip RabbitMQ, orders still place and match, they just do not settle into your
positions. If you skip the auth secrets, signup and signin will not work but the rest of
the app does.

## Run the full stack with Tilt

Tilt builds the images, applies the Kubernetes manifests in `deployments/k8s`, and runs
the database migrations through init containers.

```
tilt up
```

Open the Tilt UI to watch each service come up. The web app is on
[localhost:3000](http://localhost:3000) and the gateway on
[localhost:8080](http://localhost:8080).

## Run pieces by hand

Each Go service is a normal main package. From the repo root:

```
go run ./services/user
go run ./services/order
go run ./services/mrktdata
go run ./services/auth
go run ./gateway
```

The orderbook is a Rust binary:

```
cargo run -p order_book
```

The web app:

```
cd web && pnpm install && pnpm dev
```

You still need Postgres, Redis, and RabbitMQ running, and the migrations applied. See
[tools.md](tools.md) for running migrations with goose.

## Build and test

```
go build ./... && go test ./...     # all Go services, from the repo root
make proto                          # regenerate gRPC stubs after editing proto
make test                           # Go, Rust, and web unit tests
cargo test -p order_book            # just the matching engine
cd web && pnpm test && pnpm build   # web tests and the full type check
```

## Demo mode

`GLYPH_TRADING_247=1` fills orders even when the market is closed, which is useful for
demos. `make resetData` wipes user data back to a clean slate, so only run it when you
actually want that.
