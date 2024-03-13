package api

import (
	"os"
	"testing"
	"time"

	"github.com/anil1226/go-simplebank-grpc/store"
	"github.com/anil1226/go-simplebank-grpc/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func NewTestServer(t *testing.T, store store.Store) *Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store)
	require.NoError(t, err)
	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
