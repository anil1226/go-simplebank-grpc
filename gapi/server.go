package gapi

import (
	"fmt"

	"github.com/anil1226/go-simplebank-grpc/pb"
	"github.com/anil1226/go-simplebank-grpc/store"
	"github.com/anil1226/go-simplebank-grpc/token"
	"github.com/anil1226/go-simplebank-grpc/util"
)

type Server struct {
	pb.UnimplementedSimpleBankServer
	config     util.Config
	store      store.Store
	tokenMaker token.Maker
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

	return server, nil
}
