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

func TestLoginErrors(t *testing.T) {
	t.Run("returns internal error when session cannot be saved", func(t *testing.T) {
		repository := &fakeAuthRepository{upsertErr: stderrors.New("database error")}
		service := NewAuthService(repository)

		_, apiErr := service.Login(
			context.Background(),
			dto.LoginRequest{Username: "tiago", Password: "right-password"},
			dto.LoginSessionInfo{IPAddress: "127.0.0.1"},
		)

		if apiErr == nil {
			t.Fatal("Expected create session error")
		}
		if apiErr.GetStatus() != http.StatusInternalServerError {
			t.Fatalf("Expected status %d, got %d", http.StatusInternalServerError, apiErr.GetStatus())
		}
	})
}

func TestLoginCreatesCSRFToken(t *testing.T) {
	repository := &fakeAuthRepository{}
	service := NewAuthService(repository)

	response, apiErr := service.Login(
		context.Background(),
		dto.LoginRequest{Username: "tiago", Password: "right-password"},
		dto.LoginSessionInfo{IPAddress: "127.0.0.1"},
	)

	if apiErr != nil {
		t.Fatalf("Expected successful login, got status %d", apiErr.GetStatus())
	}
	if response.CSRFToken == "" {
		t.Fatal("Expected CSRF token to be returned for cookie transport")
	}
	if repository.upsertSession.CSRFHash == "" {
		t.Fatal("Expected CSRF hash to be persisted in session")
	}
	if repository.upsertSession.CSRFHash == response.CSRFToken {
		t.Fatal("Expected CSRF token to be hashed before persistence")
	}
}

func TestValidateSession(t *testing.T) {
	t.Run("returns user and touches session", func(t *testing.T) {
		repository := &fakeAuthRepository{
			session: models.UserSession{
				ID:   10,
				User: models.User{ID: 1, Username: "tiago"},
			},
		}
		service := NewAuthService(repository)

		user, apiErr := service.ValidateSession(context.Background(), "session-token")

		if apiErr != nil {
			t.Fatalf("Expected valid session, got status %d", apiErr.GetStatus())
		}
		if user.ID != 1 {
			t.Fatalf("Expected user ID 1, got %d", user.ID)
		}
		if !repository.findSessionCalled {
			t.Fatal("Expected session lookup to be called")
		}
		if repository.touchedSessionID != 10 {
			t.Fatalf("Expected touched session ID 10, got %d", repository.touchedSessionID)
		}
	})

	t.Run("returns unauthorized when session is invalid", func(t *testing.T) {
		repository := &fakeAuthRepository{findSessionErr: stderrors.New("not found")}
		service := NewAuthService(repository)

		_, apiErr := service.ValidateSession(context.Background(), "bad-token")

		if apiErr == nil {
			t.Fatal("Expected invalid session error")
		}
		if apiErr.GetStatus() != http.StatusUnauthorized {
			t.Fatalf("Expected status %d, got %d", http.StatusUnauthorized, apiErr.GetStatus())
		}
		if repository.touchedSessionID != 0 {
			t.Fatalf("Expected session not to be touched, got %d", repository.touchedSessionID)
		}
	})

	t.Run("returns internal error when touch fails", func(t *testing.T) {
		repository := &fakeAuthRepository{
			session: models.UserSession{
				ID:   10,
				User: models.User{ID: 1, Username: "tiago"},
			},
			touchErr: stderrors.New("database error"),
		}
		service := NewAuthService(repository)

		_, apiErr := service.ValidateSession(context.Background(), "session-token")

		if apiErr == nil {
			t.Fatal("Expected touch session error")
		}
		if apiErr.GetStatus() != http.StatusInternalServerError {
			t.Fatalf("Expected status %d, got %d", http.StatusInternalServerError, apiErr.GetStatus())
		}
	})
}

func TestValidateCSRF(t *testing.T) {
	t.Run("accepts matching csrf token for valid session", func(t *testing.T) {
		repository := &fakeAuthRepository{
			session: models.UserSession{CSRFHash: hashSessionToken("csrf-token")},
		}
		service := NewAuthService(repository)

		apiErr := service.ValidateCSRF(context.Background(), "session-token", "csrf-token")

		if apiErr != nil {
			t.Fatalf("Expected valid CSRF token, got status %d", apiErr.GetStatus())
		}
	})

	t.Run("rejects missing session token", func(t *testing.T) {
		repository := &fakeAuthRepository{}
		service := NewAuthService(repository)

		apiErr := service.ValidateCSRF(context.Background(), "", "csrf-token")

		if apiErr == nil {
			t.Fatal("Expected invalid CSRF error")
		}
		if apiErr.GetStatus() != http.StatusForbidden {
			t.Fatalf("Expected status %d, got %d", http.StatusForbidden, apiErr.GetStatus())
		}
		if repository.findSessionCalled {
			t.Fatal("Expected session lookup not to be called")
		}
	})

	t.Run("rejects missing csrf token", func(t *testing.T) {
		repository := &fakeAuthRepository{}
		service := NewAuthService(repository)

		apiErr := service.ValidateCSRF(context.Background(), "session-token", "")

		if apiErr == nil {
			t.Fatal("Expected invalid CSRF error")
		}
		if apiErr.GetStatus() != http.StatusForbidden {
			t.Fatalf("Expected status %d, got %d", http.StatusForbidden, apiErr.GetStatus())
		}
	})

	t.Run("rejects invalid session", func(t *testing.T) {
		repository := &fakeAuthRepository{findSessionErr: stderrors.New("not found")}
		service := NewAuthService(repository)

		apiErr := service.ValidateCSRF(context.Background(), "session-token", "csrf-token")

		if apiErr == nil {
			t.Fatal("Expected invalid CSRF error")
		}
		if apiErr.GetStatus() != http.StatusForbidden {
			t.Fatalf("Expected status %d, got %d", http.StatusForbidden, apiErr.GetStatus())
		}
	})

	t.Run("rejects mismatched csrf token", func(t *testing.T) {
		repository := &fakeAuthRepository{
			session: models.UserSession{CSRFHash: hashSessionToken("other-token")},
		}
		service := NewAuthService(repository)

		apiErr := service.ValidateCSRF(context.Background(), "session-token", "csrf-token")

		if apiErr == nil {
			t.Fatal("Expected invalid CSRF error")
		}
		if apiErr.GetStatus() != http.StatusForbidden {
			t.Fatalf("Expected status %d, got %d", http.StatusForbidden, apiErr.GetStatus())
		}
	})
}

func TestLogout(t *testing.T) {
	t.Run("rejects empty token", func(t *testing.T) {
		repository := &fakeAuthRepository{}
		service := NewAuthService(repository)

		apiErr := service.Logout(context.Background(), "")

		if apiErr == nil {
			t.Fatal("Expected invalid session error")
		}
		if apiErr.GetStatus() != http.StatusUnauthorized {
			t.Fatalf("Expected status %d, got %d", http.StatusUnauthorized, apiErr.GetStatus())
		}
		if repository.revokeCalled {
			t.Fatal("Expected revoke not to be called for empty token")
		}
	})

	t.Run("revokes session", func(t *testing.T) {
		repository := &fakeAuthRepository{}
		service := NewAuthService(repository)

		apiErr := service.Logout(context.Background(), "session-token")

		if apiErr != nil {
			t.Fatalf("Expected successful logout, got status %d", apiErr.GetStatus())
		}
		if !repository.revokeCalled {
			t.Fatal("Expected session revoke to be called")
		}
		if repository.revokedTokenHash == "" {
			t.Fatal("Expected token hash to be passed to repository")
		}
	})

	t.Run("returns internal error when revoke fails", func(t *testing.T) {
		repository := &fakeAuthRepository{revokeErr: stderrors.New("database error")}
		service := NewAuthService(repository)

		apiErr := service.Logout(context.Background(), "session-token")

		if apiErr == nil {
			t.Fatal("Expected logout error")
		}
		if apiErr.GetStatus() != http.StatusInternalServerError {
			t.Fatalf("Expected status %d, got %d", http.StatusInternalServerError, apiErr.GetStatus())
		}
	})
}

func TestLoginAttemptKey(t *testing.T) {
	got := loginAttemptKey(" 127.0.0.1 ", " TiAgO ")
	want := "127.0.0.1|tiago"
	if got != want {
		t.Fatalf("Expected key %q, got %q", want, got)
	}
}

func TestLoginAttemptLimiterResetsExpiredWindow(t *testing.T) {
	limiter := newLoginAttemptLimiter()
	now := time.Now().UTC()
	key := "127.0.0.1|tiago"

	limiter.RecordFailure(key, now.Add(-loginAttemptWindow-time.Second))

	if limiter.IsBlocked(key, now) {
		t.Fatal("Expected expired window not to be blocked")
	}
	if _, exists := limiter.attempts[key]; exists {
		t.Fatal("Expected expired login attempt window to be removed")
	}
}

func TestLoginAttemptLimiterStartsNewWindowAfterExpiration(t *testing.T) {
	limiter := newLoginAttemptLimiter()
	firstAttempt := time.Now().UTC()
	key := "127.0.0.1|tiago"

	limiter.RecordFailure(key, firstAttempt)
	limiter.RecordFailure(key, firstAttempt.Add(loginAttemptWindow+time.Second))

	attempt := limiter.attempts[key]
	if attempt.Count != 1 {
		t.Fatalf("Expected new window count to be 1, got %d", attempt.Count)
	}
}

type fakeAuthRepository struct {
	authenticateCalls int
	authenticateErr   error
	findSessionCalled bool
	findSessionErr    error
	revokeCalled      bool
	revokeErr         error
	revokedTokenHash  string
	session           models.UserSession
	touchErr          error
	touchedSessionID  uint
	upsertErr         error
	upsertSession     models.UserSession
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
	f.findSessionCalled = true
	if f.findSessionErr != nil {
		return models.UserSession{}, f.findSessionErr
	}
	return f.session, nil
}

func (f *fakeAuthRepository) RevokeSessionByTokenHash(_ context.Context, tokenHash string, _ time.Time) error {
	f.revokeCalled = true
	f.revokedTokenHash = tokenHash
	return f.revokeErr
}

func (f *fakeAuthRepository) TouchSession(_ context.Context, sessionID uint, _ time.Time) error {
	f.touchedSessionID = sessionID
	return f.touchErr
}

func (f *fakeAuthRepository) UpsertSession(_ context.Context, session models.UserSession) error {
	f.upsertSession = session
	return f.upsertErr
}
