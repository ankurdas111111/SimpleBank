package api

import (
	"fmt"
	"net/http"

	db "github.com/ankurdas111111/simplebank/db/sqlc"
	"github.com/ankurdas111111/simplebank/token"
	"github.com/ankurdas111111/simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	config util.Config
	store db.Store
	tokenMaker token.Maker
	router *gin.Engine
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	// tokenMaker, err := token.NewJWTMaker(config.TokenSymmetricKey)
	if err != nil{
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		config: config,
		store: store,
		tokenMaker: tokenMaker,
	}
	
	if v,ok := binding.Validator.Engine().(*validator.Validate); ok{
		v.RegisterValidation("currency",validCurrency)
	}

	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	// API (preferred): /api/*
	apiRoutes := router.Group("/api")
	apiRoutes.POST("/users", server.createUser)
	apiRoutes.POST("/users/login", server.loginUser)

	// Backward-compatible routes (older clients): keep these too.
	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))
	apiAuthRoutes := router.Group("/api").Use(authMiddleware(server.tokenMaker))

	authRoutes.POST("/accounts", server.createAccount)
	authRoutes.GET("/accounts/:id", server.getAccount)
	authRoutes.GET("/accounts", server.listAccount)
	authRoutes.POST("/accounts/:id/deposit", server.deposit)
	authRoutes.GET("/accounts/:id/lookup", server.lookupAccount)

	authRoutes.POST("/transfers", server.createTransfer)
	authRoutes.GET("/transfers", server.listTransfers)

	apiAuthRoutes.POST("/accounts", server.createAccount)
	apiAuthRoutes.GET("/accounts/:id", server.getAccount)
	apiAuthRoutes.GET("/accounts", server.listAccount)
	apiAuthRoutes.POST("/accounts/:id/deposit", server.deposit)
	apiAuthRoutes.GET("/accounts/:id/lookup", server.lookupAccount)
	apiAuthRoutes.POST("/transfers", server.createTransfer)
	apiAuthRoutes.GET("/transfers", server.listTransfers)

	// UI (served by backend for single-service deploy)
	router.GET("/", func(ctx *gin.Context) { ctx.File("./web/index.html") })
	router.GET("/dashboard.html", func(ctx *gin.Context) { ctx.File("./web/dashboard.html") })
	router.GET("/login.html", func(ctx *gin.Context) { ctx.File("./web/login.html") })
	router.GET("/signup.html", func(ctx *gin.Context) { ctx.File("./web/signup.html") })
	router.GET("/accounts.html", func(ctx *gin.Context) { ctx.File("./web/accounts.html") })
	router.GET("/transfers.html", func(ctx *gin.Context) { ctx.File("./web/transfers.html") })
	router.GET("/app.js", func(ctx *gin.Context) { ctx.File("./web/app.js") })
	router.GET("/styles.css", func(ctx *gin.Context) { ctx.File("./web/styles.css") })
	router.GET("/favicon.ico", func(ctx *gin.Context) { ctx.Status(http.StatusNoContent) })

	server.router = router
}

// start runs the server on a specific address 
func (server *Server) Start(address string) error{
	return server.router.Run(address) 
}

func errorResponse(err error) gin.H{
	return gin.H{"error": err.Error()}
}
