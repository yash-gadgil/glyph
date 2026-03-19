# Order endpoints

The order service exposes one gRPC service, `OrderService`, defined in `proto/order.proto`.
The gateway calls it for the order routes, and the user service strategy engine calls
`PlaceOrder` when a deployment fires. Money fields are integer cents.

| RPC | Request to response | Description |
|-----|---------------------|-------------|
| `PlaceOrder` | `PlaceOrderRequest` to `PlaceOrderResponse` | validate, reserve through the user service, persist, forward to the orderbook |
| `CancelOrder` | `CancelOrderRequest` to `CancelOrderResponse` | cancel in the database and in the book |
| `GetOrders` | `GetOrdersRequest` to `GetOrdersResponse` | list a user's orders, optional status filter |
| `GetOrder` | `GetOrderRequest` to `Order` | one order |
| `GetFills` | `GetFillsRequest` to `GetFillsResponse` | a user's fills |
| `GetStrategyFills` | `GetStrategyFillsRequest` to `GetFillsResponse` | the fills made by one strategy |
| `UpdateOrderStatus` | `UpdateOrderStatusRequest` to `UpdateOrderStatusResponse` | internal status update |
