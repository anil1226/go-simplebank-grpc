package api

import (
	"github.com/anil1226/go-simplebank-grpc/store"
	"github.com/gin-gonic/gin"
)

type Server struct {
	store  store.Store
	router *gin.Engine
}

func NewServer(store store.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccount)
	server.router = router

	return server
}

func (s *Server) Start(address string) error {
	return s.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
