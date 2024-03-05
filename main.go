package main

import (
	"database/sql"
	"log"

	"github.com/anil1226/go-simplebank-grpc/api"
	"github.com/anil1226/go-simplebank-grpc/store"
	_ "github.com/lib/pq"
)

const (
	dbDriver      = "postgres"
	dbSource      = "postgresql://root:secret@localhost:5432/postgres?sslmode=disable"
	serverAddress = "0.0.0.0:8080"
)

func main() {

	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("not able to connect to db")
	}

	store := store.NewStore(conn)

	server := api.NewServer(store)

	err = server.Start(serverAddress)
	if err != nil {
		log.Fatal("cannot start server")
	}

}
