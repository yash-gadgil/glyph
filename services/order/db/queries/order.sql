-- name: CreateOrder :one
INSERT INTO orders (user_id, symbol, side, order_type, time_in_force, qty, price, stop_price, status, strategy_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetOrderById :one
SELECT * FROM orders
WHERE id = $1;

-- name: GetOrdersByUserAndStatus :many
SELECT * FROM orders
WHERE user_id = $1 AND status = $2
  AND NOT (status = 5 AND updated_at < now() - interval '5 minutes')
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetOrdersByUser :many
SELECT * FROM orders
WHERE user_id = $1
  AND NOT (status IN (4, 5) AND updated_at < now() - interval '5 minutes')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateOrderStatus :exec
UPDATE orders
SET status = $2, updated_at = now()
WHERE id = $1;

-- name: UpdateOrderFill :exec
UPDATE orders
SET filled_qty = $2, status = $3, updated_at = now()
WHERE id = $1;

-- name: ApplyFillToOrder :one
UPDATE orders
SET filled_qty = filled_qty + $2,
    status = CASE WHEN filled_qty + $2 >= qty THEN 3 ELSE 2 END,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: FinalizeOrder :one
UPDATE orders
SET status = $2, updated_at = now()
WHERE id = $1 AND status IN (0, 1, 2)
RETURNING *;

-- name: CancelOrder :exec
UPDATE orders
SET status = 4, updated_at = now()
WHERE id = $1 AND status IN (0, 1, 2);

-- name: InsertFill :execrows
INSERT INTO fills (trade_id, order_id, symbol, side, qty, price_cents, liquidity, executed_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT DO NOTHING;

-- name: GetFillsByOrder :many
SELECT * FROM fills
WHERE order_id = $1
ORDER BY executed_at DESC;

-- name: GetFillsByUser :many
SELECT f.trade_id, f.order_id, f.symbol, f.side, f.qty, f.price_cents, f.liquidity, f.executed_at
FROM fills f
JOIN orders o ON o.id = f.order_id
WHERE o.user_id = $1
ORDER BY f.executed_at DESC
LIMIT $2 OFFSET $3;

-- name: GetOpenOrderSymbols :many
SELECT DISTINCT symbol FROM orders
WHERE status IN (1, 2);

-- name: GetOpenDayOrders :many
SELECT * FROM orders
WHERE status IN (0, 1, 2) AND time_in_force = 0;

-- name: GetStrategyFills :many
SELECT f.trade_id, f.order_id, f.symbol, f.side, f.qty, f.price_cents, f.liquidity, f.executed_at
FROM fills f
JOIN orders o ON o.id = f.order_id
WHERE o.strategy_id = $1 AND o.user_id = $2
ORDER BY f.executed_at DESC
LIMIT $3 OFFSET $4;

-- name: DeleteFillsForUser :execrows
DELETE FROM fills
WHERE order_id IN (SELECT id FROM orders WHERE user_id = $1);

-- name: DeleteOrdersForUser :execrows
DELETE FROM orders
WHERE user_id = $1;
