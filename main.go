package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/anil1226/go-simplebank-grpc/api"
	_ "github.com/anil1226/go-simplebank-grpc/doc/statik"
	"github.com/anil1226/go-simplebank-grpc/gapi"
	"github.com/anil1226/go-simplebank-grpc/pb"
	"github.com/anil1226/go-simplebank-grpc/store"
	"github.com/anil1226/go-simplebank-grpc/util"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/rakyll/statik/fs"
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

	//runGinServer(config, store)
	//basicHttpServer(config, store)

}

func basicHttpServer(config util.Config, store store.Store) {
	fs := http.FileServer(http.Dir("doc/swagger/"))
	http.Handle("/swagger/", http.StripPrefix("/swagger/", fs))
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello")
	})
	http.ListenAndServe(":8090", nil)
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

	fs, err := fs.New()
	if err != nil {
		log.Fatal("cannot create statik fs")
	}
	swagHandler := http.StripPrefix("/swagger/", http.FileServer(fs))
	mux.Handle("/swagger/", swagHandler)

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
