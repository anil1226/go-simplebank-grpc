package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type renewTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type rennewTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func (s *Server) renewToken(ctx *gin.Context) {
	var req renewTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	payload, err := s.tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	sess, err := s.store.GetSession(ctx, payload.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, errorResponse(err))
		return
	}

	if sess.IsBlocked {
		err := errors.New("user blocked")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	if sess.Username != payload.Username {
		err := errors.New("incorreect user session")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	if sess.RefreshToken != req.RefreshToken {
		err := errors.New("incorreect refresh token")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	if time.Now().After(sess.ExpiresAt) {
		err := errors.New("session expired")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	accessToken, payload, err := s.tokenMaker.CreateToken(payload.Username, s.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := rennewTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: payload.ExpiredAt,
	}
	ctx.JSON(http.StatusOK, resp)
}
