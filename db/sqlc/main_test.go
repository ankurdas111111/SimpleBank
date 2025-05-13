package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/ankurdas111111/simplebank/util"
	_ "github.com/lib/pq" // PostgreSQL driver
)

var testQueries *Queries
var testDB *sql.DB
var testStore Store



// TestMain acts as the setup function for all tests in the package
func TestMain(m *testing.M) {

	config,err := util.LoadConfig("../..")
	if err != nil{
		log.Fatal("Could not connect to db:",err)
	}

	testDB, err = sql.Open(config.DBdriver, config.DBsource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testQueries = New(testDB)
	testStore = NewStore(testDB)
	
	// Run all tests and exit with the appropriate code
	os.Exit(m.Run())
}