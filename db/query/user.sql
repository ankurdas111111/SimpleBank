-- name: CreateUser :one
INSERT INTO users (
    username,
    hashed_password,
    full_name,
    email    
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetUser :one
-- Direct primary key lookup ensures O(1) performance via B-tree index
-- LIMIT 1 optimizes query planning - tells PostgreSQL to stop after first match
SELECT * FROM users
WHERE username = $1 LIMIT 1;