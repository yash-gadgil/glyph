# Glyph

A paper-trading platform with a real matching engine and an LLM trading copilot.

Glyph runs a full microservice trading stack: a Next.js frontend talks to a single Go
gateway over HTTP, WebSocket, and SSE, and the gateway fans out to backend services over
gRPC. Orders match against the live market price through a Rust matching engine, fills
settle asynchronously over RabbitMQ, and a tool-calling chat agent can look up your
portfolio, check prices, and generate deployable strategies on request.

## What it does

- **Paper trading** with a real order book. Orders reserve cash or shares, match against
  the live price (the same model Alpaca paper trading uses), and settle into positions.
- **Live market data** from Alpaca, streamed to the web app and injected into the order
  book so parked orders fill.
- **Strategies and backtests.** Build rule-based strategies (RSI, SMA, MACD, Bollinger,
  and more) in their own service, backtest them over historical bars, and deploy them to
  trade automatically on a one-minute evaluation loop.
- **Kenaz, the trading copilot.** A chat agent that calls tools to read your portfolio,
  fetch prices, movers, and news, and generate a strategy for a symbol, backtesting it
  before it's ever shown to you. Answers stream over SSE; off-topic questions get turned
  away.

## Architecture

The web app only talks to the gateway, which fans everything out over gRPC. A trade
flows gateway -> order -> user (reserve) -> order book (match) -> RabbitMQ (fill) -> user
(settle).

The order book is the one Rust service in an otherwise Go and Python stack: a real
matching engine, one book per symbol on its own thread, handling market, limit, IOC/FOK,
and stop orders. It never crosses user orders against each other, everything fills
against the live market price, the same way Alpaca's own paper trading works.

Kenaz (the chat agent), the strategy engine, and everything else hang off the gateway the
same way. See [docs/architecture.md](docs/architecture.md) for the full diagram and every
flow.

## Tech stack

- **Frontend:** Next.js, React, Tailwind, TanStack Query, lightweight-charts
- **Backend:** Go (chi, gRPC), Rust (matching engine), Python (model hosting via uv)
- **LLM:** local Qwen model via the inference service, or Gemini as a hosted alternative
- **Contracts:** Protobuf with buf, generated for Go and Python
- **Data:** Postgres (sqlc, goose migrations, one database per service), Redis
- **Messaging:** RabbitMQ
- **Infra:** Kubernetes, Tilt, Docker, Prometheus, Grafana

## Quickstart

Prerequisites: Go, Rust, Node and pnpm, Docker, a Kubernetes context (Docker Desktop is
fine), Tilt, and the build tools in [docs/tools.md](docs/tools.md) (buf, sqlc, goose).
You also need free Alpaca paper-trading keys for market data.

```bash
cp .env.example .env   # fill in Alpaca keys and secrets
tilt up                # builds images, applies k8s manifests, runs migrations
```

The web app comes up on [localhost:3000](http://localhost:3000) and the gateway on
[localhost:8080](http://localhost:8080). Full instructions, including running services by
hand, are in [docs/SETUP.md](docs/SETUP.md).

The first `tilt up` downloads the advisor's local model once into a persistent volume (the
inference init container), so the first start takes a few minutes; later starts are fast.

## Docs

- [Architecture](docs/architecture.md)
- [Setup](docs/SETUP.md)
- [Build tools](docs/tools.md)
- Per-service references under [docs/services](docs/services)
