-- name: CreateCommission :one
INSERT INTO commissions (order_id, affiliate_id, amount) VALUES ($1, $2, $3) RETURNING *;

-- name: GetCommissionByID :one
SELECT id, order_id, affiliate_id, amount FROM commissions WHERE id = $1;
  
-- name: ListCommissions :many
SELECT id, order_id, affiliate_id, amount FROM commissions;

-- name: GetCommissionByOrderID :many
SELECT a.id, a.name, c.amount 
FROM commissions c 
JOIN affiliates a ON c.affiliate_id = a.id 
WHERE c.order_id = $1;

-- name: GetTotalCommission :one
SELECT SUM(amount)::FLOAT FROM commissions WHERE order_id = $1;