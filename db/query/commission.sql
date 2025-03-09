-- name: CreateCommission :one
INSERT INTO commissions (order_id, affiliate_id, amount) VALUES ($1, $2, $3) RETURNING *;

-- name: GetCommissionByID :one
SELECT id, order_id, affiliate_id, amount FROM commissions WHERE id = $1;
  
-- name: ListCommissions :many
SELECT id, order_id, affiliate_id, amount FROM commissions;