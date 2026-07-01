# Setup

This gets the whole stack running locally. There are two ways to run it: Tilt for the
full Kubernetes stack, or by hand if you only need a few pieces.

## Prerequisites

- Go (see the version in `go.mod`)
- Rust (stable) for the orderbook
- Python and [uv](https://github.com/astral-sh/uv) for the inference service
- Node and pnpm for the web app
- Docker
- The extra build tools in [tools.md](tools.md): goose, sqlc, buf, and the protoc Go plugins
- An Alpaca account for market data keys (the paper trading keys are free)
- A Gemini API key if you want the advisor to use Gemini instead of the local model

## Environment

Copy the example env file and fill it in:

```
cp .env.example .env
```

You need, at a minimum:

- `APCA_API_KEY_ID` and `APCA_API_SECRET_KEY` for Alpaca market data
- the service ports (the defaults in `.env.example` are fine)
- the Postgres connection settings for `userdb`, `orderdb`, and `strategydb`
- `RABBITMQ_URL` if you want fills to settle and account deletions to cascade to
  strategies
- the auth secrets (JWT keys, Google OAuth, SMTP) only if you are working on auth
- `GEMINI_API_KEY` only if you want the advisor's `gemini` provider; without it the
  `inference` provider (the local model) still works

If you skip RabbitMQ, orders still place and match, they just do not settle into your
positions, and deleting an account won't clean up its strategies. If you skip the auth
secrets, signup and signin will not work but the rest of the app does.

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
go run ./services/strategy
go run ./services/advisor
go run ./services/gateway
```

The orderbook is a Rust binary:

```
cargo run -p order_book
```

The inference service is a Python process, run through uv. It needs the generated Python
protobuf stubs on its path, and the `model` extra if you want to actually load a model
rather than fall back to an echo responder:

```
cd services/inference
uv sync --extra model
PYTHONPATH=../gen/python uv run python server.py
```

Without `MODEL_ID` set it falls back to the echo responder, which is enough to exercise
the advisor without a multi-minute model download.

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
demos. `./scripts/reset-data.sh [namespace]` wipes `userdb` and `orderdb` back to a clean
slate and flushes the auth cache, so only run it when you actually want that.
