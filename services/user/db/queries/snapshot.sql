-- name: InsertSnapshot :exec
INSERT INTO account_value_snapshots (user_id, equity_cents, cash_cents, market_value_cents)
VALUES ($1, $2, $3, $4);

-- name: GetSnapshotsSince :many
SELECT user_id, equity_cents, cash_cents, market_value_cents, captured_at
FROM account_value_snapshots
WHERE user_id = $1 AND captured_at >= $2
ORDER BY captured_at ASC;

-- name: PruneSnapshots :execrows
DELETE FROM account_value_snapshots
WHERE captured_at < $1;

-- name: GetAllAccounts :many
SELECT user_id, cash_balance
FROM accounts;

-- name: GetAllOpenPositions :many
SELECT user_id, symbol, qty, cost_basis
FROM positions
WHERE qty != 0;
