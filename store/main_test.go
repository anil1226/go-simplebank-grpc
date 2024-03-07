package store

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/anil1226/go-simplebank-grpc/util"
	_ "github.com/lib/pq"
)

var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error
	config, err := util.LoadConfig("../")
	if err != nil {
		log.Fatal("not able to load config")
	}
	testDB, err = sql.Open(config.DbDriver, config.DbSource)
	if err != nil {
		log.Fatal("not able to connect to db")
	}

	testQueries = New(testDB)

	os.Exit(m.Run())
}
