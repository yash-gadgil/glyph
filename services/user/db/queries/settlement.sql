-- name: InsertSettlement :execrows
INSERT INTO settlements (trade_id, order_id, user_id, symbol, side, qty, price_cents, cash_delta_cents, realized_pnl_cents)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT DO NOTHING;

-- name: GetSettlementsForUser :many
SELECT trade_id, order_id, symbol, side, qty, price_cents, cash_delta_cents, realized_pnl_cents, applied_at
FROM settlements
WHERE user_id = $1
ORDER BY applied_at DESC
LIMIT $2 OFFSET $3;

-- name: DeleteSettlementsForUser :exec
DELETE FROM settlements
WHERE user_id = $1;
