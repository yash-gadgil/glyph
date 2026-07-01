# Gateway

The gateway is the only thing the web app talks to. It is a Go HTTP server built on chi
that terminates auth, WebSocket, and SSE connections from the browser and fans requests
out to the backend services over gRPC. Nothing in the browser talks to a backend service
directly.

Path: `services/gateway/` · Port: 8080 · Depends on: every backend service (auth, user,
mrktdata, order, strategy, advisor).

## Structure

Each resource area is a route group registered in `server/handlers`: auth, account,
portfolio, orders, watchlists, explore, strategies, advisor. A group holds a thin gRPC
client for whichever service backs it and translates HTTP requests into RPC calls,
converting cents and enums into the shapes the frontend expects.

## Auth

The JWT cookie middleware runs on every route except `/auth`. It verifies the access
token against the auth service's public keys, cached by key id for up to an hour and
refreshed through `GetPublicKeys` on a miss, and injects the user id into the request
context so every downstream handler can assume it's there.

## Streaming

Two transports carry live data:

- **WebSocket** for watchlist price streaming (`/watchlists/{id}` and
  `/watchlists/stream`), upgraded with gorilla/websocket and authenticated from the cookie
  inside the handler.
- **SSE** for the advisor chat (`POST /advisor/chat`), which proxies the advisor's
  streaming `ChatWithAdvisor` RPC a chunk at a time and flushes after every write so the
  browser sees tokens as they arrive.

## Middleware

Every request passes through CORS, structured request logging, and a per-route rate
limiter before it reaches a handler. See [endpoints.md](endpoints.md) for the full route
list.
