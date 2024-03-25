package worker

import (
	"context"

	"github.com/anil1226/go-simplebank-grpc/store"
	"github.com/hibiken/asynq"
)

type TaskProcessor interface {
	Start() error
	ProcessTaskSendVerifyEmail(
		ctx context.Context,
		task *asynq.Task,
	) error
}

type RedisTaskProcessor struct {
	server *asynq.Server
	store  store.Store
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store store.Store) TaskProcessor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{},
	)
	return &RedisTaskProcessor{
		server: server,
		store:  store,
	}
}

func (rtp *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskSendVerifyEmail, rtp.ProcessTaskSendVerifyEmail)

	return rtp.server.Start(mux)
}
