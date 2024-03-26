package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const TaskSendVerifyEmail = "task:send_verify_email"

type PayLoadSendVerifyEmail struct {
	Username string `json:"username"`
}

func (rtd *RedisTaskDistributor) DistributeTaskSendVerifyEmail(
	ctx context.Context,
	payload *PayLoadSendVerifyEmail,
	opts ...asynq.Option,
) error {
	plByte, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(TaskSendVerifyEmail, plByte, opts...)
	taskInfo, err := rtd.client.EnqueueContext(ctx, task)
	if err != nil {
		return err
	}
	log.Info().
		Str("type", taskInfo.Type).
		Bytes("payload", task.Payload()).
		Str("queue", taskInfo.Queue).
		Int("max_retry", taskInfo.MaxRetry).
		Msg("enqueue task")
	return nil
}

func (rtp *RedisTaskProcessor) ProcessTaskSendVerifyEmail(
	ctx context.Context,
	task *asynq.Task,
) error {
	var payload PayLoadSendVerifyEmail
	err := json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}
	user, err := rtp.store.GetUser(ctx, payload.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user does not exist: %w", asynq.SkipRetry)
		}
		return fmt.Errorf("failed to get user: %w", err)
	}
	log.Info().
		Str("type", task.Type()).
		Bytes("payload", task.Payload()).
		Str("email", user.Email).
		Msg("process task")
	//todo: send email
	return nil
}
