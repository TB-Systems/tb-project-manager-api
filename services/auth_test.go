package services

import (
	"context"
	stderrors "errors"
	"net/http"
	"testing"
	"time"

	"github.com/TB-Systems/tb-project-manager-api/dto"
	"github.com/TB-Systems/tb-project-manager-api/models"
)

func TestLoginRateLimit(t *testing.T) {
	t.Run("blocks repeated failed login attempts for same IP and username", func(t *testing.T) {
		repository := &fakeAuthRepository{authenticateErr: stderrors.New("invalid credentials")}
		service := NewAuthService(repository)
		request := dto.LoginRequest{Username: "tiago", Password: "wrong-password"}
		sessionInfo := dto.LoginSessionInfo{IPAddress: "127.0.0.1"}

		for i := 0; i < maxFailedLoginAttempts; i++ {
			_, apiErr := service.Login(context.Background(), request, sessionInfo)
			if apiErr == nil {
				t.Fatal("Expected invalid credentials error")
			}
			if apiErr.GetStatus() != http.StatusUnauthorized {
				t.Fatalf("Expected status %d, got %d", http.StatusUnauthorized, apiErr.GetStatus())
			}
		}

		_, apiErr := service.Login(context.Background(), request, sessionInfo)
		if apiErr == nil {
			t.Fatal("Expected rate limit error")
		}
		if apiErr.GetStatus() != http.StatusTooManyRequests {
			t.Fatalf("Expected status %d, got %d", http.StatusTooManyRequests, apiErr.GetStatus())
		}
		if repository.authenticateCalls != maxFailedLoginAttempts {
			t.Errorf("Expected %d authenticate calls, got %d", maxFailedLoginAttempts, repository.authenticateCalls)
		}
	})

	t.Run("successful login resets failed attempts", func(t *testing.T) {
		repository := &fakeAuthRepository{authenticateErr: stderrors.New("invalid credentials")}
		service := NewAuthService(repository)
		request := dto.LoginRequest{Username: "tiago", Password: "wrong-password"}
		sessionInfo := dto.LoginSessionInfo{IPAddress: "127.0.0.1"}

		for i := 0; i < maxFailedLoginAttempts-1; i++ {
			_, apiErr := service.Login(context.Background(), request, sessionInfo)
			if apiErr == nil || apiErr.GetStatus() != http.StatusUnauthorized {
				t.Fatal("Expected invalid credentials before successful login")
			}
		}

		repository.authenticateErr = nil
		_, apiErr := service.Login(context.Background(), dto.LoginRequest{Username: "tiago", Password: "right-password"}, sessionInfo)
		if apiErr != nil {
			t.Fatalf("Expected successful login, got status %d", apiErr.GetStatus())
		}

		repository.authenticateErr = stderrors.New("invalid credentials")
		_, apiErr = service.Login(context.Background(), request, sessionInfo)
		if apiErr == nil {
			t.Fatal("Expected invalid credentials after reset")
		}
		if apiErr.GetStatus() != http.StatusUnauthorized {
			t.Fatalf("Expected status %d after reset, got %d", http.StatusUnauthorized, apiErr.GetStatus())
		}
	})
}

type fakeAuthRepository struct {
	authenticateCalls int
	authenticateErr   error
	upsertErr         error
	user              models.User
}

func (f *fakeAuthRepository) Authenticate(context.Context, models.Login) (models.User, error) {
	f.authenticateCalls++
	if f.authenticateErr != nil {
		return models.User{}, f.authenticateErr
	}
	if f.user.ID == 0 {
		f.user = models.User{
			ID:       1,
			Name:     "Tiago",
			Username: "tiago",
			Email:    "tiago@example.com",
			CPF:      "00000000000",
		}
	}
	return f.user, nil
}

func (f *fakeAuthRepository) FindValidSessionByTokenHash(context.Context, string, time.Time) (models.UserSession, error) {
	return models.UserSession{}, nil
}

func (f *fakeAuthRepository) RevokeSessionByTokenHash(context.Context, string, time.Time) error {
	return nil
}

func (f *fakeAuthRepository) TouchSession(context.Context, uint, time.Time) error {
	return nil
}

func (f *fakeAuthRepository) UpsertSession(context.Context, models.UserSession) error {
	return f.upsertErr
}
