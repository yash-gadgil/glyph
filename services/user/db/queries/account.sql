-- name: GetPortfolio :one
SELECT cash_balance, reserved_cash, currency, multiplier, margin_used
FROM accounts
WHERE user_id = $1;

-- name: GetAccountByUser :one
SELECT * FROM accounts
WHERE user_id = $1;

-- name: CreateAccount :one
INSERT into accounts (user_id)
VALUES ($1)
RETURNING id;

-- name: EnsureAccount :exec
INSERT INTO accounts (user_id)
VALUES ($1)
ON CONFLICT (user_id) DO NOTHING;

-- name: AddFunds :one
UPDATE accounts
SET cash_balance = cash_balance + $2
WHERE user_id = $1
RETURNING cash_balance;

-- name: ResetAccountBalances :exec
UPDATE accounts
SET cash_balance = 10000000, reserved_cash = 0, margin_used = 0
WHERE user_id = $1;

-- name: ReserveCash :one
UPDATE accounts
SET reserved_cash = reserved_cash + $2
WHERE user_id = $1 AND cash_balance - reserved_cash >= $2
RETURNING reserved_cash;

-- name: ReleaseCash :exec
UPDATE accounts
SET reserved_cash = GREATEST(reserved_cash - $2, 0)
WHERE user_id = $1;

-- name: ApplyCashDelta :exec
UPDATE accounts
SET cash_balance = cash_balance + $2,
    reserved_cash = GREATEST(reserved_cash - $3, 0)
WHERE user_id = $1;
