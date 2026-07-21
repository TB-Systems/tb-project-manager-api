package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/tb-project-manager-api/dto"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/TB-Systems/tb-project-manager-api/repositories"
)

type Auth interface {
	Login(ctx context.Context, request dto.LoginRequest, sessionInfo dto.LoginSessionInfo) (dto.LoginResponse, errors.ApiError)
	Logout(ctx context.Context, token string) errors.ApiError
	Session(ctx context.Context, token string) (dto.AuthSessionResponse, errors.ApiError)
	ValidateCSRF(ctx context.Context, sessionToken string, csrfToken string) errors.ApiError
	ValidateSession(ctx context.Context, token string) (models.User, errors.ApiError)
}

type auth struct {
	repository          repositories.Auth
	loginAttemptLimiter *loginAttemptLimiter
}

func NewAuthService(repository repositories.Auth) Auth {
	return &auth{
		repository:          repository,
		loginAttemptLimiter: newLoginAttemptLimiter(),
	}
}

const (
	sessionTTL             = 24 * time.Hour
	maxFailedLoginAttempts = 5
	loginAttemptWindow     = 5 * time.Minute
	loginAttemptBlockTTL   = 5 * time.Minute
)

func (a *auth) Login(ctx context.Context, request dto.LoginRequest, sessionInfo dto.LoginSessionInfo) (dto.LoginResponse, errors.ApiError) {
	login := models.Login{
		Username: request.Username,
		Password: request.Password,
	}

	attemptKey := loginAttemptKey(sessionInfo.IPAddress, request.Username)
	if a.loginAttemptLimiter.IsBlocked(attemptKey, time.Now().UTC()) {
		return dto.LoginResponse{}, errors.NewApiError(
			http.StatusTooManyRequests,
			errors.BadRequestError("TOO_MANY_LOGIN_ATTEMPTS"),
		)
	}

	user, err := a.repository.Authenticate(ctx, login)
	if err != nil {
		a.loginAttemptLimiter.RecordFailure(attemptKey, time.Now().UTC())
		return dto.LoginResponse{}, errors.NewApiError(
			http.StatusUnauthorized,
			errors.BadRequestError("INVALID_CREDENTIALS"),
		)
	}

	a.loginAttemptLimiter.Reset(attemptKey)

	token, tokenHash, err := newSessionToken()
	if err != nil {
		return dto.LoginResponse{}, errors.NewApiError(
			http.StatusInternalServerError,
			errors.InternalServerError("CREATE_SESSION_TOKEN_FAILED"),
		)
	}

	csrfToken, csrfHash, err := newSessionToken()
	if err != nil {
		return dto.LoginResponse{}, errors.NewApiError(
			http.StatusInternalServerError,
			errors.InternalServerError("CREATE_CSRF_TOKEN_FAILED"),
		)
	}

	now := time.Now().UTC()
	expiresAt := now.Add(sessionTTL)
	session := models.UserSession{
		UserID:     user.ID,
		TokenHash:  tokenHash,
		CSRFHash:   csrfHash,
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
		CSRFToken:    csrfToken,
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

func (a *auth) ValidateSession(ctx context.Context, token string) (models.User, errors.ApiError) {
	session, apiErr := a.validSession(ctx, token)
	if apiErr != nil {
		return models.User{}, apiErr
	}

	return session.User, nil
}

func (a *auth) Session(ctx context.Context, token string) (dto.AuthSessionResponse, errors.ApiError) {
	session, apiErr := a.validSession(ctx, token)
	if apiErr != nil {
		return dto.AuthSessionResponse{}, apiErr
	}

	return dto.AuthSessionResponse{
		ID:        session.User.ID,
		Name:      session.User.Name,
		Username:  session.User.Username,
		Email:     session.User.Email,
		CPF:       session.User.CPF,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

func (a *auth) validSession(ctx context.Context, token string) (models.UserSession, errors.ApiError) {
	now := time.Now().UTC()
	session, err := a.repository.FindValidSessionByTokenHash(ctx, hashSessionToken(token), now)
	if err != nil {
		return models.UserSession{}, errors.NewApiError(
			http.StatusUnauthorized,
			errors.BadRequestError("INVALID_SESSION"),
		)
	}

	if err := a.repository.TouchSession(ctx, session.ID, now); err != nil {
		return models.UserSession{}, errors.NewApiError(
			http.StatusInternalServerError,
			errors.InternalServerError("UPDATE_SESSION_FAILED"),
		)
	}

	return session, nil
}

func (a *auth) Logout(ctx context.Context, token string) errors.ApiError {
	if token == "" {
		return errors.NewApiError(
			http.StatusUnauthorized,
			errors.BadRequestError("INVALID_SESSION"),
		)
	}

	if err := a.repository.RevokeSessionByTokenHash(ctx, hashSessionToken(token), time.Now().UTC()); err != nil {
		return errors.NewApiError(
			http.StatusInternalServerError,
			errors.InternalServerError("LOGOUT_FAILED"),
		)
	}

	return nil
}

func (a *auth) ValidateCSRF(ctx context.Context, sessionToken string, csrfToken string) errors.ApiError {
	if sessionToken == "" || csrfToken == "" {
		return invalidCSRFError()
	}

	session, err := a.repository.FindValidSessionByTokenHash(ctx, hashSessionToken(sessionToken), time.Now().UTC())
	if err != nil {
		return invalidCSRFError()
	}

	if session.CSRFHash == "" || session.CSRFHash != hashSessionToken(csrfToken) {
		return invalidCSRFError()
	}

	return nil
}

func invalidCSRFError() errors.ApiError {
	return errors.NewApiError(
		http.StatusForbidden,
		errors.BadRequestError("INVALID_CSRF_TOKEN"),
	)
}

func hashSessionToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

type loginAttemptLimiter struct {
	mu       sync.Mutex
	attempts map[string]loginAttempt
}

type loginAttempt struct {
	Count         int
	FirstFailedAt time.Time
	BlockedUntil  time.Time
}

func newLoginAttemptLimiter() *loginAttemptLimiter {
	return &loginAttemptLimiter{attempts: make(map[string]loginAttempt)}
}

func (l *loginAttemptLimiter) IsBlocked(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	attempt, exists := l.attempts[key]
	if !exists {
		return false
	}

	if attempt.BlockedUntil.After(now) {
		return true
	}

	if now.Sub(attempt.FirstFailedAt) > loginAttemptWindow {
		delete(l.attempts, key)
	}

	return false
}

func (l *loginAttemptLimiter) RecordFailure(key string, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()

	attempt := l.attempts[key]
	if attempt.FirstFailedAt.IsZero() || now.Sub(attempt.FirstFailedAt) > loginAttemptWindow {
		attempt = loginAttempt{FirstFailedAt: now}
	}

	attempt.Count++
	if attempt.Count >= maxFailedLoginAttempts {
		attempt.BlockedUntil = now.Add(loginAttemptBlockTTL)
	}

	l.attempts[key] = attempt
}

func (l *loginAttemptLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.attempts, key)
}

func loginAttemptKey(ipAddress string, username string) string {
	return strings.TrimSpace(ipAddress) + "|" + strings.ToLower(strings.TrimSpace(username))
}
