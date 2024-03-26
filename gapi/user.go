package gapi

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/anil1226/go-simplebank-grpc/pb"
	"github.com/anil1226/go-simplebank-grpc/store"
	"github.com/anil1226/go-simplebank-grpc/util"
	"github.com/anil1226/go-simplebank-grpc/val"
	"github.com/anil1226/go-simplebank-grpc/worker"
	"github.com/hibiken/asynq"
	"github.com/lib/pq"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) CreateUser(ctx context.Context, in *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {

	errs := validateCreateUserRequest(in)
	if errs != nil {
		return nil, invalidArgumentError(errs)
	}
	hPassword, err := util.HashedPassword(in.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	arg := store.CreateUserTxParams{
		CreateUserParams: store.CreateUserParams{
			Username:       in.Username,
			HashedPassword: hPassword,
			FullName:       in.FullName,
			Email:          in.Email,
		},
		AfterCreate: func(user store.User) error {
			taskPayload := &worker.PayLoadSendVerifyEmail{
				Username: user.Username,
			}

			opts := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(worker.QueueCritical),
			}
			return s.taskDistributor.DistributeTaskSendVerifyEmail(ctx, taskPayload, opts...)

		},
	}

	usr, err := s.store.CreateUserTx(ctx, arg)
	if err != nil {
		if pq, ok := err.(*pq.Error); ok {
			log.Println(pq.Code.Name())
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateUserResponse{
		User: &pb.User{
			Username:          usr.User.Username,
			FullName:          usr.User.FullName,
			Email:             usr.User.Email,
			PasswordChangedAt: timestamppb.New(usr.User.PasswordChangedAt),
			CreatedAt:         timestamppb.New(usr.User.CreatedAt),
		},
	}, nil
}
func (s *Server) LoginUser(ctx context.Context, in *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	errs := validateLoginUserRequest(in)
	if errs != nil {
		return nil, invalidArgumentError(errs)
	}
	user, err := s.store.GetUser(ctx, in.Username)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	err = util.CheckPassword(in.Password, user.HashedPassword)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	accessToken, payload, err := s.tokenMaker.CreateToken(in.Username, s.config.AccessTokenDuration)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(in.Username, s.config.RefreshTokenDuration)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	meta := s.extractMetaData(ctx)

	session, err := s.store.CreateSession(ctx, store.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     in.Username,
		RefreshToken: refreshToken,
		UserAgent:    meta.UserAgent,
		ClientIp:     meta.ClientIP,
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &pb.LoginUserResponse{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  timestamppb.New(payload.ExpiredAt),
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: timestamppb.New(refreshPayload.ExpiredAt),
		SessionId:             session.ID.String(),
		User: &pb.User{
			Username: user.Username,
			FullName: user.FullName,
		},
	}
	return resp, nil
}

func validateCreateUserRequest(in *pb.CreateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(in.Username); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}
	if err := val.ValidatePassword(in.Password); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}
	if err := val.ValidateEmail(in.Email); err != nil {
		violations = append(violations, fieldViolation("email", err))
	}
	if err := val.ValidateFullname(in.FullName); err != nil {
		violations = append(violations, fieldViolation("fullname", err))
	}
	return
}

func validateLoginUserRequest(in *pb.LoginUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(in.Username); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}
	if err := val.ValidatePassword(in.Password); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}
	return
}

func validateUpdateUserRequest(in *pb.UpdateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(in.Username); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}
	if in.Password != nil {
		if err := val.ValidatePassword(*in.Password); err != nil {
			violations = append(violations, fieldViolation("password", err))
		}
	}
	if in.Email != nil {
		if err := val.ValidateEmail(*in.Email); err != nil {
			violations = append(violations, fieldViolation("email", err))
		}
	}
	if in.FullName != nil {
		if err := val.ValidateFullname(*in.FullName); err != nil {
			violations = append(violations, fieldViolation("fullname", err))
		}
	}
	return
}

func (s *Server) UpdateUser(ctx context.Context, in *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {

	payload, err := s.authorizeUser(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	errs := validateUpdateUserRequest(in)
	if errs != nil {
		return nil, invalidArgumentError(errs)
	}

	if payload.Username != in.Username {
		return nil, status.Error(codes.PermissionDenied, "user not allowed")
	}

	arg := store.UpdateUserParams{
		Username: in.Username,
		FullName: sql.NullString{
			String: in.GetFullName(),
			Valid:  in.FullName != nil,
		},
		Email: sql.NullString{
			String: in.GetEmail(),
			Valid:  in.Email != nil,
		},
	}

	if in.Password != nil {
		hPassword, err := util.HashedPassword(*in.Password)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		arg.HashedPassword = sql.NullString{
			String: hPassword,
			Valid:  true,
		}
		arg.PasswordChangedAt = sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}
	}

	usr, err := s.store.UpdateUser(ctx, arg)
	if err != nil {
		if pq, ok := err.(*pq.Error); ok {
			log.Println(pq.Code.Name())
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.UpdateUserResponse{
		User: &pb.User{
			Username:          usr.Username,
			FullName:          usr.FullName,
			Email:             usr.Email,
			PasswordChangedAt: timestamppb.New(usr.PasswordChangedAt),
			CreatedAt:         timestamppb.New(usr.CreatedAt),
		},
	}, nil
}
