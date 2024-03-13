package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const minSecretKeySize = 32

type JWTMaker struct {
	secretKey string
}

func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size")
	}
	return &JWTMaker{secretKey: secretKey}, nil
}

type MyCustomClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	claims := MyCustomClaims{
		payload.Username,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(payload.ExpiredAt),
			IssuedAt:  jwt.NewNumericDate(payload.IssuedAt),
			ID:        payload.ID.String(),
		},
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return jwtToken.SignedString([]byte(maker.secretKey))
}
func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	keyfunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("token is invalid")
		}
		return []byte(maker.secretKey), nil
	}
	jwttoken, err := jwt.ParseWithClaims(token, &MyCustomClaims{}, keyfunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, jwt.ErrTokenExpired
		}
		return nil, jwt.ErrTokenInvalidClaims
	}

	payload, ok := jwttoken.Claims.(*MyCustomClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return &Payload{
		Username:  payload.Username,
		IssuedAt:  payload.IssuedAt.Time,
		ExpiredAt: payload.ExpiresAt.Time,
	}, nil
}
