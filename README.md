# Simple Bank - Go Implementation Reference

A banking system demonstrating idiomatic Go patterns, concurrency management, and database transaction handling.

## Engineering Design

SimpleBank implements core banking functionality to showcase Go's approach to:
- Database interaction with compile-time SQL validation
- Concurrency-safe transaction processing
- Deadlock prevention via resource ordering
- Optimistic and pessimistic locking strategies
- Deterministic integration testing

## Architecture

```
SimpleBank/
├── db/
│   ├── migration/          # Schema versioning with forward/rollback capabilities
│   ├── query/              # Raw SQL with typed annotations for code generation
│   └── sqlc/               # Compile-time verified database access layer
├── util/                   # Stateless utility functions
├── makefile                # Automated build pipeline components
├── sqlc.yaml               # Code generation configuration
├── go.mod                  # Explicit dependency versioning (Go modules)
└── go.sum                  # Dependency integrity verification
```

## Technical Implementation

### Database Access Layer

The database layer implements the Repository pattern using code generation:
- **Query parameterization** for SQL injection prevention
- **Type-safe return values** mapped to Go structs
- **Compile-time SQL validation** catches errors before runtime
- **Context propagation** for request cancellation and timeout handling

### Transaction Management

`store.go` demonstrates Go's approach to transaction handling:

1. **Higher-order function pattern** for transaction execution
   ```go
   func execTx(ctx context.Context, fn func(*Queries) error) error {
       tx, err := db.BeginTx(ctx, nil)
       // Execute fn with transaction scope
       // Handle commit/rollback 
   }
   ```

2. **Resource ordering** for deadlock prevention (Coffman condition avoidance)
   ```go
   // Always process accounts in consistent ID order
   if fromID < toID {
       // Update fromAccount then toAccount
   } else {
       // Update toAccount then fromAccount
   }
   ```

3. **Atomic operations** for race-condition avoidance
   ```sql
   -- Atomic update protected from race conditions
   UPDATE accounts SET balance = balance + $2 WHERE id = $1
   ```

### Concurrent Testing Implementation

The project demonstrates Go's approach to testing concurrent systems:

1. **Goroutines** for parallel execution testing
   ```go
   for i := 0; i < n; i++ {
       go func() {
           // Test operation in parallel
       }()
   }
   ```

2. **Buffered channels** for synchronized result collection
   ```go
   results := make(chan Result, n)  // Buffered to prevent goroutine leaks
   
   // Send results from goroutines
   results <- result
   
   // Collect all results regardless of execution order
   for i := 0; i < n; i++ {
       result := <-results  // Deterministic collection count
   }
   ```

3. **Deterministic state validation** after non-deterministic execution
   ```go
   // Verify final state after all concurrent operations complete
   require.Equal(t, initialBalance - n*amount, finalBalance)
   ```

## Go Memory Model Considerations

### Value Semantics vs. Pointer Semantics

The codebase demonstrates appropriate use of Go's value semantics:

```go
// Return value semantics for immutable result
func (q *Queries) GetAccount(ctx context.Context, id int64) (Account, error) {
    // Return copy of account (not reference)
}

// Pointer receiver for query execution capabilities
func (q *Queries) CreateAccount(ctx context.Context, arg CreateAccountParams) (Account, error) {
    // q is pointer receiver for access to connection
}
```

### Zero Allocation Strategies

```go
// strings.Builder for zero-allocation string construction
func RandomString(n int) string {
    var sb strings.Builder
    // Single allocation at return point
    return sb.String()
}

// Stack allocation for fixed-size arrays
currencies := []string{"USD", "EUR", "INR"}
```

## Concurrency Patterns

### Resource Ordering

The primary mechanism for deadlock prevention uses Coffman condition avoidance:

```go
// Ensure global ordering of resource acquisition
if id1 < id2 {
    // Lock in ascending ID order
} else {
    // Still lock in ascending ID order
}
```

### Atomic Database Operations

```sql
-- Atomic operation prevents lost updates
UPDATE accounts SET balance = balance + $2 WHERE id = $1
```

### Channel-based Synchronization

```go
// Buffered channels prevent goroutine leaks
results := make(chan Result, n)

// Collect exact number of results
for i := 0; i < n; i++ {
    result := <-results
}
```

## Go Design Patterns

### Functional Options

```go
// Higher-order function for transaction execution
func execTx(ctx context.Context, fn func(*Queries) error) error {
    // Function composition pattern
}
```

### Dependency Injection

```go
// Explicit dependency injection
func NewStore(db *sql.DB) *Store {
    return &Store{
        db: db,
        Queries: New(db), 
    }
}
```

### Repository Pattern

```go
// Store combines raw DB access with query methods
type Store struct {
    db *sql.DB
    *Queries // Embedded type for method promotion
}
```

## Performance Considerations

1. **Query efficiency**
   - Primary key lookups for O(1) access
   - LIMIT clauses for early termination
   - Parameterized queries for plan caching

2. **Memory management**
   - Value semantics for immutable data
   - Pointer semantics for mutable state
   - Preallocated buffers for string operations

3. **Concurrency control**
   - Explicit locking order to prevent deadlocks
   - FOR UPDATE clauses for pessimistic locking
   - Atomic operations for consistency

## Development Environment

### Prerequisites

- Go 1.14+ (for error wrapping and module support)
- PostgreSQL 12+ (for improved performance and features)
- Docker (optional, for containerized PostgreSQL)

### Build Pipeline

```sh
# Start PostgreSQL container
make postgres

# Create database with proper encoding and collation
make createdb

# Apply schema migrations
make migrateup

# Generate type-safe query code
make sqlc

# Execute integration tests
go test ./... -v
```

## Core Go Concepts

### Interface Implementation

Go uses implicit interface satisfaction, demonstrated by database transaction handling:

```go
// DBTX interface abstracts both *sql.DB and *sql.Tx
type DBTX interface {
    ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
    PrepareContext(context.Context, string) (*sql.Stmt, error)
    QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
    QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}
```

### Error Handling Strategy

```go
// Explicit error propagation
result, err := store.TransferTx(ctx, arg)
if err != nil {
    // Handle error at appropriate level
}

// Error wrapping for context preservation
if rbErr := tx.Rollback(); rbErr != nil {
    return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
}
```

### Context Propagation

```go
// Pass context through all DB operations for cancellation support
result, err := store.TransferTx(context.Background(), params)
```

## Project Overview

SimpleBank is a Go application that provides core banking functionalities:
- Creating and managing accounts
- Recording account entries (deposits/withdrawals)
- Transferring money between accounts

The project demonstrates several important Go concepts:
- Structured SQL database access
- Transactions and concurrency handling
- Unit testing
- Database migration

## Project Structure

```
SimpleBank/
├── db/
│   ├── migration/          # Database migration files
│   ├── query/              # SQL queries
│   └── sqlc/               # Generated Go code for database access
├── util/                   # Utility functions
├── makefile                # Build and development commands
├── sqlc.yaml               # SQLC configuration
├── go.mod                  # Go module dependencies
└── go.sum                  # Dependency checksums
```

## Key Components Explained

### 1. Database Structure

The database has several key tables:
- `accounts`: Stores account information (owner, balance, currency)
- `entries`: Records individual account activities (deposits/withdrawals)
- `transfers`: Records money transfers between accounts

### 2. SQLC Code Generation

We use [SQLC](https://sqlc.dev/) to generate type-safe Go code from SQL queries, which:
- Eliminates manual SQL string building
- Provides compile-time SQL validation
- Generates appropriate Go types for each query

### 3. Database Transaction Handling

The `store.go` file implements key transaction patterns:
- A generic transaction executor (`execTx`)
- Atomic money transfers between accounts
- Deadlock prevention in concurrent transfers

### 4. Unit Testing

The project includes comprehensive tests that demonstrate:
- Individual account CRUD operations testing
- Concurrent money transfer testing with goroutines
- Random test data generation

## Core Go Concepts Demonstrated

### 1. Struct Embedding

```go
type Store struct {
    db *sql.DB
    *Queries        // Embedded struct gives direct access to query methods
}
```

Struct embedding lets us access the embedded struct's methods directly.

### 2. Goroutines and Channels

```go
// Launch concurrent transfers
for i := 0; i < n; i++ {
    go func() {
        // Do transfer
        results <- result  // Send result to channel
    }()
}

// Collect results
for i := 0; i < n; i++ {
    result := <-results    // Receive from channel
}
```

We use goroutines to run transfers concurrently and channels to collect results.

### 3. Database Transactions

```go
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
    tx, err := store.db.BeginTx(ctx, nil)
    // Execute transaction logic
    // Commit or rollback
}
```

This pattern allows us to run multiple database operations in a single transaction.

## Getting Started

### Prerequisites

- Go 1.14 or later
- PostgreSQL 12 or later
- Docker (optional, for running PostgreSQL)

### Setup

1. **Start the database**:
   ```
   make postgres
   ```

2. **Create the database**:
   ```
   make createdb
   ```

3. **Run database migrations**:
   ```
   make migrateup
   ```

4. **Generate database code**:
   ```
   make sqlc
   ```

5. **Run tests**:
   ```
   go test ./...
   ```

## Understanding the Money Transfer Logic

The money transfer process involves several steps:

1. Begin a database transaction
2. Create a transfer record
3. Create account entries for both accounts
4. Update account balances in a consistent order to prevent deadlocks
5. Commit the transaction

This ensures that money transfers are atomic - either the entire operation succeeds, or none of it does.

## Learn More About Go

- [Go Tour](https://tour.golang.org/) - Official Go tutorial
- [Go by Example](https://gobyexample.com/) - Practical Go programming examples
- [SQLC Documentation](https://docs.sqlc.dev/) - Learn about SQL code generation
- [Go Database Tutorial](https://go.dev/doc/tutorial/database-access) - Go's database access tutorial

## Common Go Concepts for Beginners

### Interfaces

Go interfaces define behavior rather than structure:

```go
type DBTX interface {
    ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
    PrepareContext(context.Context, string) (*sql.Stmt, error)
    QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
    QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}
```

### Error Handling

Go's explicit error handling approach:

```go
result, err := store.TransferTx(...)
if err != nil {
    // Handle error
}
```

### Context

Context is used for cancellation, deadlines, and passing values:

```go
result, err := store.TransferTx(context.Background(), TransferTxParams{...})
``` 