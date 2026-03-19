# Orderbook endpoints

The orderbook exposes one gRPC service, `OrderbookService`, defined in
`proto/order_book.proto`. The order service calls `AddOrder` and `CancelOrder`, and the
market data service calls `InjectPrice`. Prices and quantities are integers.

Each call is routed to the worker thread for its symbol, which owns that symbol's book.

| RPC | Request to response | Description |
|-----|---------------------|-------------|
| `AddOrder` | `AddOrderRequest` to `AddOrderResponse` | submit an order for a symbol, returns the trades it produced and whether it was accepted |
| `CancelOrder` | `CancelOrderRequest` to `CancelOrderResponse` | cancel a resting, parked, or stop order |
| `InjectPrice` | `InjectPriceRequest` to `InjectPriceResponse` | push a live price, fills parked markets, crossed limits, and triggered stops, returns the number of fills |

The order type, side, and time in force are enums in the proto. A non zero `stop_price` on
an order marks it as a stop or stop limit, which parks until the trigger trades.

Fills and terminal events are published to RabbitMQ as a side effect of these calls, not
returned to the caller. See the [overview](overview.md) for the event shape.
