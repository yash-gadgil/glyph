# Market data

The market data service is the bridge to Alpaca. It serves historical bars, the catalog of
tradable symbols, latest prices, a live price stream, and news and movers. It is a Go gRPC
service.

Path: `services/mrktdata/` · Port: 50052 · Depends on: Alpaca (the `APCA_*` keys).

## What it does

It serves historical bars, the catalog of tradable symbols, latest prices, a live price
stream, and news and movers, all sourced from Alpaca. See [endpoints.md](endpoints.md) for
the RPCs.

## Live stream

The live stream keeps a single shared connection to Alpaca and reference counts the symbol
subscriptions across all the clients watching them. When the last client stops watching a
symbol it unsubscribes, and when nothing is being watched it disconnects, so the service
holds at most one upstream connection no matter how many browsers are open.

## Feeding the orderbook

Market data is also what makes paper fills happen. It pushes the latest price for a symbol
into the orderbook through `InjectPrice`, and that tick is what fills parked market orders,
crossed limits, and triggered stops. Without these injected prices the matching engine has
nothing to fill against.
