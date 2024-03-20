package gapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/anil1226/go-simplebank-grpc/token"
	"google.golang.org/grpc/metadata"
)

func (s *Server) authorizeUser(ctx context.Context) (*token.Payload, error) {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}

	authValues := meta.Get("authorization")
	if len(authValues) == 0 {
		return nil, fmt.Errorf("missing auth header")
	}
	fields := strings.Fields(authValues[0])
	if len(fields) < 2 {
		return nil, fmt.Errorf("empty auth header")
	}

	authType := strings.ToLower(fields[0])
	if authType != "bearer" {
		return nil, fmt.Errorf("auth header not bearer")
	}

	accessToken := fields[1]
	payLoad, err := s.tokenMaker.VerifyToken(accessToken)
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}
	return payLoad, nil
}
