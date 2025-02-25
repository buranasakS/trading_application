-- name: CreateProduct :one
INSERT INTO products (name, quantity, price) VALUES ($1, $2, $3) RETURNING * ;

-- name: ListProducts :many
SELECT id, name, quantity, price FROM products;

-- name: GetProductByID :one
SELECT id, name, quantity, price FROM products WHERE id = $1;

-- name: CreateAffiliate :one
INSERT INTO affiliates (name, master_affiliate, balance)
VALUES ($1, $2, 0)
RETURNING *;

-- name: ListAffiliates :many
SELECT id, name, master_affiliate, balance FROM affiliates;

-- name: GetAffiliateByID :one
SELECT id, name, master_affiliate, balance FROM affiliates WHERE id = $1;

-- name: CreateCommission :one
INSERT INTO commissions (order_id, affiliate_id, amount) VALUES ($1, $2, $3) RETURNING *;

-- name: GetCommissionByID :one
SELECT id, order_id, affiliate_id, amount FROM commissions WHERE id = $1;
  
-- name: ListCommissions :many
SELECT id, order_id, affiliate_id, amount FROM commissions;

-- name: CreateUser :one
INSERT INTO users (username, affiliate_id) VALUES ($1, $2) RETURNING *;

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

-- name: AddUserBalance :exec
UPDATE users SET balance = balance + $1 WHERE id = $2;

-- name: CheckUserExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE id = $1);

-- name: UserBalance :one
SELECT id, balance FROM users WHERE id = $1;

-- name: DeductProductQuantity :execrows
UPDATE products SET quantity = quantity - $1 WHERE id = $2 AND quantity >= $1;

-- name: GetAffiliateByUserID :one
SELECT id, master_affiliate FROM affiliates WHERE id = $1;

-- name: AddAffiliateBalance :exec
UPDATE affiliates SET balance = balance + $1 WHERE id = $2;