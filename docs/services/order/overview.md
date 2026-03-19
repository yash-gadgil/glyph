# Order

The order service runs the order lifecycle: validate, reserve, persist, forward to the
matching engine, and track the result. It is a Go gRPC service backed by Postgres
(`orderdb`) through sqlc.

Path: `services/order/` · Port: 50055 · Depends on: Postgres, user service, orderbook,
RabbitMQ.

## What it does

- validates an order by its type (market, limit, IOC, FOK, stop, stop limit)
- reserves cash or shares by calling the user service before the order goes anywhere
- persists the order in `orderdb`
- forwards the order to the orderbook over gRPC and records the trades that come back
- cancels orders both in the database and in the book

## The place order flow

When you place an order the service first calls `ReserveForOrder` on the user service. If
you cannot afford the buy or do not hold the shares to sell, the order is rejected right
there and never reaches the matching engine. Once the reservation succeeds the order is
saved and forwarded to the orderbook.

The actual settlement does not happen here. The orderbook publishes fill events to
RabbitMQ, and the user service consumes them to update positions and cash. Keeping
settlement on the message path means a fill is applied exactly once even if the same order
is touched by more than one request.

## Endpoints

See [endpoints.md](endpoints.md) for the full `OrderService` RPC list.

## Database

The schema, queries, and generated code live under `db/`. As with the other services, sqlc
reads `schema.sql` and migrations run with goose.
