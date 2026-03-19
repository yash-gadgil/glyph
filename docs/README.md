# Glyph docs

Glyph is a paper trading platform. You get a cash balance, place orders against live
market data, run strategies, and watch how your portfolio does over time. Nothing uses
real money.

The backend is a set of Go microservices with a Rust matching engine. The frontend is a
Next.js web app. Everything talks to a single Go gateway over HTTP and WebSocket, and the
gateway fans out to the backend services over gRPC.

All money is stored as integer cents. There are no floats anywhere in the money path.

## Where to start

- [SETUP.md](SETUP.md) gets the whole stack running on your machine
- [tools.md](tools.md) covers the extra command line tools the build needs
- [architecture.md](architecture.md) is the big picture: services, data flow, ports

## Services

Each service has its own page under [services/](services/):

- [gateway](services/gateway/overview.md) the public HTTP and WebSocket edge
- [auth](services/auth/overview.md) signup, signin, OAuth, JWTs, email
- [user](services/user/overview.md) accounts, portfolio, watchlists, strategies
- [mrktdata](services/mrktdata/overview.md) market data from Alpaca
- [order](services/order/overview.md) the order lifecycle
- [orderbook](services/orderbook/overview.md) the Rust matching engine

## Frontend

The web app is the only client. It lives outside the backend services and just talks to
the gateway.

- [web](web/overview.md) the Next.js frontend
