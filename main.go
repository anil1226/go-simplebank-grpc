package main

import (
	"database/sql"
	"log"

	"github.com/anil1226/go-simplebank-grpc/api"
	"github.com/anil1226/go-simplebank-grpc/store"
	"github.com/anil1226/go-simplebank-grpc/util"
	_ "github.com/lib/pq"
)

func main() {

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("not able to load config")
	}

	conn, err := sql.Open(config.DbDriver, config.DbSource)
	if err != nil {
		log.Fatal("not able to connect to db")
	}

	store := store.NewStore(conn)

	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server")
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server")
	}

}
