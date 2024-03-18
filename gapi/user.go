package gapi

import (
	"context"
	"log"

	"github.com/anil1226/go-simplebank-grpc/pb"
	"github.com/anil1226/go-simplebank-grpc/store"
	"github.com/anil1226/go-simplebank-grpc/util"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) CreateUser(ctx context.Context, in *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	hPassword, err := util.HashedPassword(in.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	arg := store.CreateUserParams{
		Username:       in.Username,
		HashedPassword: hPassword,
		FullName:       in.FullName,
		Email:          in.Email,
	}

	usr, err := s.store.CreateUser(ctx, arg)
	if err != nil {
		if pq, ok := err.(*pq.Error); ok {
			log.Println(pq.Code.Name())
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.CreateUserResponse{
		User: &pb.User{
			Username:          usr.Username,
			FullName:          usr.FullName,
			Email:             usr.Email,
			PasswordChangedAt: timestamppb.New(usr.PasswordChangedAt),
			CreatedAt:         timestamppb.New(usr.CreatedAt),
		},
	}, nil
}
func (s *Server) LoginUser(ctx context.Context, in *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
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

	session, err := s.store.CreateSession(ctx, store.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     in.Username,
		RefreshToken: refreshToken,
		UserAgent:    "",
		ClientIp:     "",
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
