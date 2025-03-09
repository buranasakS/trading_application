-- name: CreateUser :one
INSERT INTO users (username, password, affiliate_id) VALUES ($1, $2, $3) RETURNING *;

-- name: ListUsers :many
SELECT id, username, balance, affiliate_id
FROM users
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: GetUserDetailByID :one
SELECT id, username, balance, affiliate_id FROM users WHERE id = $1;

-- name: DeductUserBalance :execrows
UPDATE users SET balance = balance - $1 WHERE id = $2 AND balance >= $1;

-- name: AddUserBalance :execrows
UPDATE users SET balance = balance + $1 WHERE id = $2;

-- name: CheckUserExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE id = $1);

-- name: UserBalance :one
SELECT id, balance FROM users WHERE id = $1;

-- name: GetUserByUsernameForLogin :one
SELECT id, username, password, affiliate_id, balance
FROM users
WHERE username = $1
LIMIT 1;