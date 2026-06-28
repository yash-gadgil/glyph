-- name: CreateUser :one
INSERT INTO users (email, password_hash, user_name)
VALUES ($1, $2, $3)
RETURNING id;

-- name: CheckEmailAvailability :one
SELECT EXISTS (
  SELECT 1
  FROM users
  WHERE email = $1
);

-- name: GetUserPassword :one
SELECT id, password_hash
FROM users
WHERE email = $1;

-- name: UpdateUserPasswordByEmail :exec
UPDATE users SET password_hash = $1 WHERE email = $2;

-- name: UpdateUserPasswordById :exec
UPDATE users SET password_hash = $1 WHERE id = $2;

-- name: GetUserById :one
SELECT id, email, user_name FROM users
WHERE id = $1;

-- name: DeleteUser :execrows
DELETE FROM users
WHERE id = $1;
