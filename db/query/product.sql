-- name: CreateProduct :one
INSERT INTO products (name, quantity, price) VALUES ($1, $2, $3) RETURNING * ;

-- name: ListProducts :many
SELECT id, name, quantity, price FROM products;

-- name: GetProductByID :one
SELECT id, name, quantity, price FROM products WHERE id = $1;

-- name: DeductProductQuantity :execrows
UPDATE products SET quantity = quantity - $1 WHERE id = $2 AND quantity >= $1;


