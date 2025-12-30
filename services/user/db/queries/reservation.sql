-- name: CreateReservation :exec
INSERT INTO order_reservations (order_id, user_id, symbol, side, qty, remaining_qty, cents_per_share)
VALUES ($1, $2, $3, $4, $5, $5, $6);

-- name: GetReservation :one
SELECT * FROM order_reservations
WHERE order_id = $1;

-- name: GetReservationForUpdate :one
SELECT * FROM order_reservations
WHERE order_id = $1
FOR UPDATE;

-- name: ReduceReservation :one
UPDATE order_reservations
SET remaining_qty = remaining_qty - $2
WHERE order_id = $1
RETURNING *;

-- name: DeleteReservation :exec
DELETE FROM order_reservations
WHERE order_id = $1;

-- name: DeleteReservationsForUser :exec
DELETE FROM order_reservations
WHERE user_id = $1;
