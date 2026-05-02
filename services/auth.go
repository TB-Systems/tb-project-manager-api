package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/tb-project-manager-api/dto"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/TB-Systems/tb-project-manager-api/repositories"
)

type Auth interface {
	Login(ctx context.Context, request dto.LoginRequest, sessionInfo dto.LoginSessionInfo) (dto.LoginResponse, errors.ApiError)
	ValidateSession(ctx context.Context, token string) (models.User, errors.ApiError)
}

type auth struct {
	repository repositories.Auth
}

func NewAuthService(repository repositories.Auth) Auth {
	return auth{repository: repository}
}

const sessionTTL = 24 * time.Hour

func (a auth) Login(ctx context.Context, request dto.LoginRequest, sessionInfo dto.LoginSessionInfo) (dto.LoginResponse, errors.ApiError) {
	login := models.Login{
		Username: request.Username,
		Password: request.Password,
	}

	user, err := a.repository.Authenticate(ctx, login)
	if err != nil {
		return dto.LoginResponse{}, errors.NewApiError(
			http.StatusUnauthorized,
			errors.BadRequestError("INVALID_CREDENTIALS"),
		)
	}

	token, tokenHash, err := newSessionToken()
	if err != nil {
		return dto.LoginResponse{}, errors.NewApiError(
			http.StatusInternalServerError,
			errors.InternalServerError("CREATE_SESSION_TOKEN_FAILED"),
		)
	}

	now := time.Now().UTC()
	expiresAt := now.Add(sessionTTL)
	session := models.UserSession{
		UserID:     user.ID,
		TokenHash:  tokenHash,
		UserAgent:  sessionInfo.UserAgent,
		IPAddress:  sessionInfo.IPAddress,
		ExpiresAt:  expiresAt,
		LastSeenAt: now,
	}

	if err := a.repository.UpsertSession(ctx, session); err != nil {
		return dto.LoginResponse{}, errors.NewApiError(
			http.StatusInternalServerError,
			errors.InternalServerError("CREATE_SESSION_FAILED"),
		)
	}

	return dto.LoginResponse{
		ID:           user.ID,
		Name:         user.Name,
		Username:     user.Username,
		Email:        user.Email,
		CPF:          user.CPF,
		SessionToken: token,
		ExpiresAt:    expiresAt,
	}, nil
}

func newSessionToken() (string, string, error) {
	rawToken := make([]byte, 32)
	if _, err := rand.Read(rawToken); err != nil {
		return "", "", fmt.Errorf("read random token: %w", err)
	}

	token := base64.RawURLEncoding.EncodeToString(rawToken)

	return token, hashSessionToken(token), nil
}

func (a auth) ValidateSession(ctx context.Context, token string) (models.User, errors.ApiError) {
	now := time.Now().UTC()
	session, err := a.repository.FindValidSessionByTokenHash(ctx, hashSessionToken(token), now)
	if err != nil {
		return models.User{}, errors.NewApiError(
			http.StatusUnauthorized,
			errors.BadRequestError("INVALID_SESSION"),
		)
	}

	if err := a.repository.TouchSession(ctx, session.ID, now); err != nil {
		return models.User{}, errors.NewApiError(
			http.StatusInternalServerError,
			errors.InternalServerError("UPDATE_SESSION_FAILED"),
		)
	}

	return session.User, nil
}

func hashSessionToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
