package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/anil1226/go-simplebank-grpc/api"
	_ "github.com/anil1226/go-simplebank-grpc/doc/statik"
	"github.com/anil1226/go-simplebank-grpc/gapi"
	"github.com/anil1226/go-simplebank-grpc/pb"
	"github.com/anil1226/go-simplebank-grpc/store"
	"github.com/anil1226/go-simplebank-grpc/util"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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
		log.Fatal().Msg("not able to load config")
	}

	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	conn, err := sql.Open(config.DbDriver, config.DbSource)
	if err != nil {
		log.Fatal().Msg("not able to connect to db")
	}

	runDBMigration("file://migrations", config.DbSource)

	store := store.NewStore(conn)

	go runGatewayServer(config, store)

	runGRPCServer(config, store)

	//runGinServer(config, store)
	//basicHttpServer(config, store)

}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Msg("cannot create migration")
	}
	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Msg("cannot run migrate up")
	}
	log.Info().Msg("db migration successfull")
}

func basicHttpServer(config util.Config, store store.Store) {
	fs := http.FileServer(http.Dir("doc/swagger/"))
	http.Handle("/swagger/", http.StripPrefix("/swagger/", fs))
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Print(w, "hello")
	})
	http.ListenAndServe(":8090", nil)
}

func runGRPCServer(config util.Config, store store.Store) {

	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal().Msg("cannot create server")
	}
	logger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer(logger)
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	lister, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal().Msg("cannot start listerer:")
	}

	log.Info().Msgf("start gRPC server at %s", lister.Addr().String())
	err = grpcServer.Serve(lister)
	if err != nil {
		log.Fatal().Msg("cannot start server")
	}
}

func runGatewayServer(config util.Config, store store.Store) {

	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal().Msg("cannot create server")
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
		log.Fatal().Msg("cannot register handler server")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	fs, err := fs.New()
	if err != nil {
		log.Fatal().Msg("cannot create statik fs")
	}
	swagHandler := http.StripPrefix("/swagger/", http.FileServer(fs))
	mux.Handle("/swagger/", swagHandler)

	lister, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Msg("cannot start listerer")
	}

	log.Info().Msgf("start HTTP server at %s", lister.Addr().String())
	err = http.Serve(lister, mux)
	if err != nil {
		log.Fatal().Msg("cannot start server")
	}
}

func runGinServer(config util.Config, store store.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal().Msg("cannot create server")
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Msg("cannot start server")
	}
}
