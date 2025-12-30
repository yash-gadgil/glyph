-- name: GetPositionsForUser :many
SELECT symbol, qty, reserved_qty, realized_pnl, cost_basis, updated_at
FROM positions
WHERE user_id = $1
ORDER BY symbol;

-- name: GetPosition :one
SELECT symbol, qty, reserved_qty, realized_pnl, cost_basis, updated_at
FROM positions
WHERE user_id = $1 AND symbol = $2;

-- name: GetPositionForUpdate :one
SELECT symbol, qty, reserved_qty, realized_pnl, cost_basis
FROM positions
WHERE user_id = $1 AND symbol = $2
FOR UPDATE;

-- name: SetPosition :exec
INSERT INTO positions (user_id, symbol, qty, realized_pnl, cost_basis, reserved_qty, updated_at)
VALUES ($1, $2, $3, $4, $5, 0, now())
ON CONFLICT (user_id, symbol) DO UPDATE
SET qty = EXCLUDED.qty,
    realized_pnl = EXCLUDED.realized_pnl,
    cost_basis = EXCLUDED.cost_basis,
    updated_at = now();

-- name: ReserveShares :one
UPDATE positions
SET reserved_qty = reserved_qty + $3
WHERE user_id = $1 AND symbol = $2 AND qty - reserved_qty >= $3
RETURNING reserved_qty;

-- name: ReleaseShares :exec
UPDATE positions
SET reserved_qty = GREATEST(reserved_qty - $3, 0)
WHERE user_id = $1 AND symbol = $2;

-- name: DeletePositionsForUser :exec
DELETE FROM positions
WHERE user_id = $1;
