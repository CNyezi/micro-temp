-- name: CreateOrder :one
INSERT INTO orders (
    user_id, total_amount, status
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetOrder :one
SELECT * FROM orders
WHERE id = $1 LIMIT 1;

-- name: UpdateOrderStatus :one
UPDATE orders
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListOrdersByUser :many
SELECT * FROM orders
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CancelOrder :exec
UPDATE orders
SET
    status = 'cancelled',
    updated_at = NOW()
WHERE id = $1;

-- name: CreateOrderItem :one
INSERT INTO order_items (
    order_id, product_id, product_name, quantity, price
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetOrderItems :many
SELECT * FROM order_items
WHERE order_id = $1
ORDER BY created_at;