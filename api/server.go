package api

import (
	"fmt"

	"github.com/anil1226/go-simplebank-grpc/store"
	"github.com/anil1226/go-simplebank-grpc/token"
	"github.com/anil1226/go-simplebank-grpc/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	config     util.Config
	store      store.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

func NewServer(config util.Config, store store.Store) (*Server, error) {
	tokenMaker, err := token.NewPaetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("not able to create token")
	}
	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}
	router := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccount)
	router.POST("/transfers", server.createTransfer)
	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)
	server.router = router

	return server, nil
}

func (s *Server) Start(address string) error {
	return s.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
