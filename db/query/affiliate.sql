-- name: CreateAffiliate :one
INSERT INTO affiliates (name, master_affiliate, balance)
VALUES ($1, $2, 0)
RETURNING *;

-- name: ListAffiliates :many
SELECT id, name, master_affiliate, balance FROM affiliates;

-- name: GetAffiliateByID :one
SELECT id, name, master_affiliate, balance FROM affiliates WHERE id = $1;

-- name: GetAffiliateByUserID :one
SELECT id, master_affiliate FROM affiliates WHERE id = $1;

-- name: AddAffiliateBalance :exec
UPDATE affiliates SET balance = balance + $1 WHERE id = $2;