package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
}

// Store implements the Repository pattern for database access
// It uses struct embedding to inherit query methods while adding transaction capabilities
type SQLStore struct {
	db *sql.DB      // Maintains a single connection pool for DB operations
	*Queries        // Embeds query methods via composition (preferred over inheritance in Go)
}

// NewStore constructs a Store instance with dependency injection pattern
// This follows Go's preference for explicit dependencies over global state
func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db), // Uses constructor pattern rather than direct initialization
	}
}

// execTx implements the functional options pattern for transaction execution
// This higher-order function accepts a function parameter for execution within a tx context
// (Higher-order functions are a key Go idiom for extending behavior)
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	// BeginTx accepts a context for propagating cancellation and deadlines
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err // Early return pattern for error handling (preferred in Go)
	}

	// Creates a query executor scoped to this transaction
	q := New(tx)
	
	// Execute the callback, maintaining the error in local scope
	err = fn(q)
	
	// Uses deferred execution via explicit error handling rather than defer
	// This provides more granular control over the transaction outcome
	if err != nil {
		// Handles nested error with wrapping for error context preservation
		if rbErr := tx.Rollback(); rbErr != nil {
			// fmt.Errorf with %v verb for error interpolation (Go 1.13+ idiom)
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}
	
	// Explicit commit required in Go (no auto-commit like some ORMs)
	return tx.Commit()
}

// TransferTxParams uses struct field tags for JSON serialization
// The json tags enable zero-allocation marshaling via reflection
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"` // Uses lowercase+underscore naming for external representation
	ToAccountID   int64 `json:"to_account_id"`   // But keeps CamelCase for Go identifiers (idiomatic Go style)
	Amount        int64 `json:"amount"`          // Uses int64 for precise currency representation (avoid float)
}

// TransferTxResult uses value semantics for immutable return data
// Go prefers returning values over mutation when possible
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`     
	FromAccount Account  `json:"from_account"` 
	ToAccount   Account  `json:"to_account"`   
	FromEntry   Entry    `json:"from_entry"`   
	ToEntry     Entry    `json:"to_entry"`     
}

// addAccountsForUpdate demonstrates the multi-value return idiom in Go
// It returns multiple values with named return parameters, which also initialize the zero value
func (store *SQLStore) addAccountsForUpdate(ctx context.Context, q *Queries, accountID1, accountID2 int64) (account1, account2 Account, err error) {
	// Swap algorithm to ensure consistent lock ordering (deadlock prevention)
	if accountID1 > accountID2 {
		accountID1, accountID2 = accountID2, accountID1 // Multiple assignment in a single statement
	}
	
	// Zero-value initialized returns are filled in sequence
	account1, err = store.getAccountForUpdate(ctx, q, accountID1)
	if err != nil {
		return // Naked return uses pre-declared return values
	}
	
	account2, err = store.getAccountForUpdate(ctx, q, accountID2)
	return // Implicit return of named return values
}

// getAccountForUpdate uses direct SQL execution rather than generated queries
// Demonstrates manual row scanning when auto-generated code isn't sufficient
func (store *SQLStore) getAccountForUpdate(ctx context.Context, q *Queries, accountID int64) (Account, error) {
	// Raw SQL string for custom locking behavior
	// The FOR UPDATE clause is DB-specific and not abstracted by the query generator
	query := `SELECT id, owner, balance, currency, created_at FROM accounts
		WHERE id = $1 LIMIT 1
		FOR UPDATE`
	
	// QueryRowContext returns at most one row and is optimized for this case
	row := q.db.QueryRowContext(ctx, query, accountID)
	
	// Using stack-allocated struct for the result (efficient for single row)
	var account Account
	// row.Scan uses the address-of operator to modify the struct in-place
	// This avoids additional allocations compared to returning a new struct
	err := row.Scan(
		&account.ID,
		&account.Owner,
		&account.Balance,
		&account.Currency,
		&account.CreatedAt,
	)
	return account, err // Return both values, error handling at the call site
}

// TransferTx demonstrates a complete transactional workflow pattern
// It uses optimistic concurrency control via SQL-level locking
func (store *SQLStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult
	
	// Uses anonymous function as a closure to capture the result variable
	// This is a common Go pattern for transactional operations
	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// Sequence of operations with chain-style error handling
		// Each operation proceeds only if previous ones succeeded
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err // Early return on failure
		}

		// Entry creation follows the same pattern
		// Note that we use negative value for outgoing money - avoids separate operation types
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount, // Unary negation operator for opposing operations
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		// Implements Coffman deadlock prevention algorithm using resource ordering
		// This is a critical pattern for concurrent systems to prevent deadlock
		if arg.FromAccountID < arg.ToAccountID {
			// Process in ID order when from < to
			result.FromAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
				ID:      arg.FromAccountID,
				Balance: -arg.Amount,
			})
			if err != nil {
				return err
			}

			result.ToAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
				ID:      arg.ToAccountID,
				Balance: arg.Amount,
			})
			if err != nil {
				return err
			}
		} else {
			// Process in reverse ID order when to < from
			// This ensures a global ordering of locks regardless of transfer direction
			result.ToAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
				ID:      arg.ToAccountID,
				Balance: arg.Amount,
			})
			if err != nil {
				return err
			}

			result.FromAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
				ID:      arg.FromAccountID,
				Balance: -arg.Amount,
			})
			if err != nil {
				return err
			}
		}

		return nil // Explicit nil return required even when error is obvious
	})

	return result, err // Return both result and error to let caller handle errors
}