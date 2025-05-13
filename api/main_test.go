package api

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)


func TestMain(m *testing.M) {

	gin.SetMode(gin.TestMode)
	// Run all tests and exit with the appropriate code
	os.Exit(m.Run())
}