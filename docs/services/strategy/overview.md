# Strategy

The strategy service owns strategy definitions, backtests, and live deployments. It used
to live inside the user service; it's now its own Go gRPC service so the evaluation engine
can tick independently of account and order traffic. It's backed by its own Postgres
database, `strategydb`, through sqlc.

Path: `services/strategy/` · Port: 50059 · Depends on: Postgres (`strategydb`), the market
data service (bars and prices), the order service (placing orders for deployments),
RabbitMQ (`user.deleted` cascade delete).

## Strategy config

A strategy is a JSON config of indicator-based rules: 1-3 entry rules and optional exit
rules, each an AND/OR group comparing an indicator (or a constant) against another
indicator with an operator (`>`, `<`, `>=`, `<=`, `crosses_above`, `crosses_below`), plus a
stop-loss and take-profit percentage. The supported indicators are price, `sma`, `ema`,
`rsi`, the three MACD lines, stochastic `%K`/`%D`, Bollinger bands, ATR, volume, and VWAP.

The same engine (`engine/`) evaluates rules for both backtests and live deployments, so a
backtest is a faithful preview of how a deployment will behave.

## Backtesting

`RunBacktest` replays a config over historical bars in memory: it simulates entries and
exits at candle closes, tracks cash and position state, and returns total return, max
drawdown, Sharpe, win rate, profit factor, number of trades, and an equity curve. Some
warmup bars are consumed by indicator lookback (for example 50 bars for an SMA(50)) before
the strategy can signal at all.

## Live deployments

`DeployStrategy` starts a deployment on a symbol with a position size. The engine ticks
once a minute, only during market hours:

1. Load every running deployment from Postgres.
2. Fetch recent bars and the latest price for each deployment's symbol from the market
   data service.
3. Evaluate entry rules if not in position, or exit rules if in position.
4. On a signal, call the order service's `PlaceOrder` (buy on entry, sell on exit) and
   update the deployment's `in_position`, entry price, and quantity.

`StopDeployment` halts the tick loop for a deployment without deleting its history;
`DeleteDeployment` removes it once stopped.

## Cascade delete

A background worker (`worker/userdeleted.go`) consumes `user.deleted` events published by
the user service on account deletion and deletes every strategy and deployment for that
user in a single transaction. Malformed events are dead-lettered; transient failures are
requeued.

## Endpoints

See [endpoints.md](endpoints.md) for the full `StrategyService` RPC list.

## Database

The schema, queries, and generated code live under `db/`. sqlc reads `schema.sql`, and
migrations run with goose, same as every other Go service.
