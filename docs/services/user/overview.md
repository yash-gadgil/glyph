# User

The user service owns identities, the trading account and its cash, positions,
watchlists, and strategies. It is the biggest backend service and it is where money
actually moves. It is a Go gRPC service backed by Postgres (`userdb`) through sqlc.

Path: `services/user/` · Port: 50053 · Depends on: Postgres, market data (for prices),
order service (for the strategy engine), RabbitMQ (for settlement).

## Structure

The gRPC handlers hold the business logic directly. Each handler file in `handlers/` is
both the gRPC server and the implementation, holding the database handle, the sqlc query
layer, and the logger. There is no separate service layer.

Four gRPC services run in the one process: account, watchlist, portfolio, and strategy.
See [endpoints.md](endpoints.md) for the RPCs each one exposes.

## Money and settlement

All money is integer cents. A paper account starts with a fixed balance, the same idea as
Alpaca paper trading, so there is no arbitrary depositing.

Two calls bracket every order. Before an order goes to the matching engine the order
service calls `ReserveForOrder`, which holds buying power for a buy or shares for a sell.
After the order is done it calls `ReleaseForOrder`, which frees whatever was not used.
Reservations are long only.

The settlement consumer is the single place that turns a fill into real position and cash
changes. It reads fill events off RabbitMQ, applies them in one transaction (a settlement
ledger insert as an idempotency gate, then the position update, then the cash movement and
hold release), and never double applies a redelivered event. The position math lives in a
pure, table tested function: buys grow the lot at cost, sells relieve cost proportionally
with banker's rounding and realize the difference.

## Workers

- the settlement consumer described above
- a snapshotter that records an account value point per account once a minute, which powers
  the portfolio chart, and works even without live prices by falling back to cost basis
- the strategy engine that evaluates running deployments on a schedule and places paper
  orders through the order service

## Strategies and backtesting

A strategy is a JSON config built in the web app. You can deploy it to run live on a symbol
or backtest it over historical bars. The strategy engine in `strategyengine/` evaluates
indicators and entry and exit rules, and the same engine drives both live deployments and
backtests, so a backtest and a live run behave the same way.

## Holdings

Holdings value your positions at the latest market price pulled from the market data
service. If a price is missing the lot is valued at cost basis so the totals stay
meaningful rather than dropping to zero.

## Database

The schema, queries, and generated code live under `db/`. sqlc reads `schema.sql`, so a
table change means editing both the migration and `schema.sql`. Migrations run with goose.
