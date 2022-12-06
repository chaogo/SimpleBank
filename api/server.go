package api

import (
	db "github.com/chaogo/SimpleBank/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Server serves HTTP requests for our banking service
type Server struct {
	store db.Store // interface, to interact with database when processing API requests from clients
	router *gin.Engine // send each API request to the correct handler for processing
}

// NewServer creates a new HTTP server and setup routing
func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	// setup routing
	router := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccount)
	
	router.POST("/transfers", server.createTransfer)

	server.router = router
	return server
}

// Start runs the HTTP server on a specific address to start listening for API request
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}


// format error to send back to the client
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}