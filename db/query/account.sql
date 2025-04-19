-- name: CreateAccount :one
-- Parameterized INSERT using positional arguments ($1, $2, $3) for SQL injection protection
-- RETURNING clause fetches newly created row in a single roundtrip, saving a subsequent SELECT
INSERT INTO accounts (
    owner,
    balance,
    currency    
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetAccount :one
-- Direct primary key lookup ensures O(1) performance via B-tree index
-- LIMIT 1 optimizes query planning - tells PostgreSQL to stop after first match
SELECT * FROM accounts
WHERE id = $1 LIMIT 1;

-- name: GetAccountForUpdate :one
-- Direct primary key lookup ensures O(1) performance via B-tree index
-- LIMIT 1 optimizes query planning - tells PostgreSQL to stop after first match
SELECT * FROM accounts
WHERE id = $1 LIMIT 1
FOR NO KEY UPDATE;

-- name: ListAccounts :many
-- Paginated query pattern with LIMIT/OFFSET for incremental data retrieval
-- ORDER BY ensures stable pagination even with concurrent modifications
-- Ordering by primary key is efficient due to clustered index usage
SELECT * FROM accounts
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateAccount :one
-- Single-row UPDATE targeting primary key for efficient index scan
-- RETURNING clause eliminates need for separate SELECT after UPDATE
-- This is an absolute-value update (overwrites existing balance)
UPDATE accounts
SET balance = $2
WHERE id = $1
RETURNING *;

-- name: UpdateAccountBalance :one
-- Atomic increment/decrement pattern for concurrent safety
-- Uses SET balance = balance + $2 for race-condition-free operation
-- Critical for maintaining consistency under concurrent modifications
UPDATE accounts
SET balance = balance + $2
WHERE id = $1
RETURNING *;

-- name: DeleteAccount :exec
-- Simple primary-key targeted DELETE operation
-- CASCADE behavior depends on foreign key constraints defined in schema
-- Returns no rows (exec) since we don't need the deleted data
DELETE FROM accounts
WHERE id = $1;
