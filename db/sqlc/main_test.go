package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq" // PostgreSQL driver
)

var testQueries *Queries
var testDB *sql.DB
var testStore *Store

// TestMain acts as the setup function for all tests in the package
func TestMain(m *testing.M) {
	// Since util.LoadConfig is likely not implemented yet, we'll use a direct connection string
	connStr := "postgresql://root:secret@localhost:5400/simple_bank?sslmode=disable"
	
	var err error
	testDB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testQueries = New(testDB)
	testStore = NewStore(testDB)
	
	// Run all tests and exit with the appropriate code
	os.Exit(m.Run())
}