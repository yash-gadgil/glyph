# Orderbook

The orderbook is the matching engine, written in Rust. It keeps the resting orders for
each symbol and fills them against the live market price. It is a separate service from the
Go order service, which sends it commands over gRPC and gets back the resulting trades.

Path: `services/order_book/` · Port: 50056 · Depends on: RabbitMQ (optional, it runs
without it and just does not publish).

All prices and quantities are integers.

## The paper trading model

This is the key idea and what makes the engine different from a real exchange. User orders
never cross each other. Every fill is against a synthetic counterparty called `market` at
the latest injected price. A buy and a sell at the same price do not trade with each other,
they each fill independently against the market price. This mirrors how Alpaca paper
trading fills against the live market rather than against other paper traders. Liquidity is
not simulated, so any quantity fills fully at the market price with no improvement beyond
it.

## What the engine handles

- market orders fill immediately at the last price, or park until the first price arrives
- limit orders fill if they are already marketable, otherwise they rest until a price tick
  crosses them
- IOC and FOK orders fill fully if marketable, otherwise they are killed untouched
- stop and stop limit orders park until the trigger trades, then release as a market or a
  resting limit order
- cancels and modifies on resting, parked, and stop orders

A price tick is the only place fills originate. One tick fills parked market orders, the
limits it crosses, and the stops it triggers, all in one outcome.

The engine lives in `src/orderbook/`. The book itself is in `book.rs`, with the order,
trade, and shared types alongside it.

## The gRPC server

Each symbol gets its own worker thread with its own book. The gRPC handler dispatches
commands to the right worker over a channel and waits for the reply, so the books are
single threaded and never shared, while different symbols run in parallel. The server code
is in `src/server/`. See [endpoints.md](endpoints.md) for the RPCs.

## Events

When an order fills or otherwise leaves the book the server publishes events to RabbitMQ:
a fill event per real order side, and a done event when an order is finished, cancelled, or
killed. The synthetic market side never settles, so it produces no event. The field names
are the contract the user service settlement consumer reads.

## Tests

The engine and the server both have unit tests, and there is an integration test suite in
`tests/` that drives the gRPC layer and the worker threads end to end.
