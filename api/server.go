package api

import (
	"fmt"

	db "github.com/ankurdas111111/simplebank/db/sqlc"
	"github.com/ankurdas111111/simplebank/token"
	"github.com/ankurdas111111/simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// server serves HTTP requests for our banking service.
type Server struct {
	config util.Config
	store db.Store
	tokenMaker token.Maker
	router *gin.Engine
}

// NewServer creates a server and setup routing
func NewServer(config util.Config, store db.Store) (*Server, error) {
		tokenMaker, err := token.NewPasetoMaker(util.RandomString(32))
	if err != nil{
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		config: config,
		store: store,
		tokenMaker: tokenMaker,
	}
	router := gin.Default()

	if v,ok := binding.Validator.Engine().(*validator.Validate); ok{
		v.RegisterValidation("currency",validCurrency)
	}

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccount)
	router.POST("/users", server.createUser)


	router.POST("/transfers", server.createTransfer)

	server.router = router
	return server, nil
}

// start runs the server on a specific address 
func (server *Server) Start(address string) error{
	return server.router.Run(address) 
}

func errorResponse(err error) gin.H{
	return gin.H{"error": err.Error()}
}
