-- name: CreateDeployment :one
INSERT INTO strategy_deployments (user_id, strategy_id, symbol, position_size_cents)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetDeployment :one
SELECT * FROM strategy_deployments
WHERE id = $1 AND user_id = $2;

-- name: GetDeploymentsForUser :many
SELECT d.*, s.name AS strategy_name
FROM strategy_deployments d
JOIN strategies s ON s.id = d.strategy_id
WHERE d.user_id = $1
ORDER BY d.created_at DESC;

-- name: GetRunningDeployments :many
SELECT d.*, s.config AS strategy_config
FROM strategy_deployments d
JOIN strategies s ON s.id = d.strategy_id
WHERE d.status = 0;

-- name: StopDeployment :one
UPDATE strategy_deployments
SET status = 1, updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: UpdateDeploymentPosition :exec
UPDATE strategy_deployments
SET in_position = $2, entry_price_cents = $3, qty = $4, updated_at = now()
WHERE id = $1;

-- name: DeleteDeployment :execrows
DELETE FROM strategy_deployments
WHERE id = $1 AND user_id = $2 AND status = 1;
