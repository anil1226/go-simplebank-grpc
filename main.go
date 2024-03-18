package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"

	"github.com/anil1226/go-simplebank-grpc/api"
	"github.com/anil1226/go-simplebank-grpc/gapi"
	"github.com/anil1226/go-simplebank-grpc/pb"
	"github.com/anil1226/go-simplebank-grpc/store"
	"github.com/anil1226/go-simplebank-grpc/util"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
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

	go runGatewayServer(config, store)

	runGRPCServer(config, store)

}

func runGRPCServer(config util.Config, store store.Store) {
	grpc := grpc.NewServer()
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server")
	}
	pb.RegisterSimpleBankServer(grpc, server)
	reflection.Register(grpc)

	lister, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal("cannot start listerer:", err)
	}

	log.Printf("start gRPC server at %s", lister.Addr().String())
	err = grpc.Serve(lister)
	if err != nil {
		log.Fatal("cannot start server")
	}
}

func runGatewayServer(config util.Config, store store.Store) {

	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server")
	}

	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	grpcMux := runtime.NewServeMux(jsonOption)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal("cannot register handler server")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	fs := http.FileServer(http.Dir("./doc/swagger"))
	mux.Handle("/swagger", http.StripPrefix("/swagger/", fs))

	lister, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot start listerer")
	}

	log.Printf("start HTTP server at %s", lister.Addr().String())
	err = http.Serve(lister, mux)
	if err != nil {
		log.Fatal("cannot start server")
	}
}

func runGinServer(config util.Config, store store.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server")
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot start server")
	}
}
