package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/tb-project-manager-api/constants"
	"github.com/TB-Systems/tb-project-manager-api/dto"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/gin-gonic/gin"
)

func TestLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns login response and sets session cookie", func(t *testing.T) {
		expiresAt := time.Now().Add(time.Hour)
		service := &fakeAuthService{
			loginResponse: dto.LoginResponse{
				ID:           1,
				Name:         "Tiago",
				Username:     "tiago",
				Email:        "tiago@example.com",
				CPF:          "00000000000",
				SessionToken: "session-token",
				CSRFToken:    "csrf-token",
				ExpiresAt:    expiresAt,
			},
		}
		router := authRouter(service)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"username":"tiago","password":"strong-password"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "test-agent")

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if !service.loginCalled {
			t.Fatal("Expected login service to be called")
		}
		if service.loginRequest.Username != "tiago" {
			t.Fatalf("Expected username %q, got %q", "tiago", service.loginRequest.Username)
		}
		if service.loginSessionInfo.UserAgent != "test-agent" {
			t.Fatalf("Expected user agent %q, got %q", "test-agent", service.loginSessionInfo.UserAgent)
		}

		cookie := findSetCookie(t, w, constants.SessionCookieName)
		if cookie.Value != "session-token" {
			t.Fatalf("Expected session cookie value %q, got %q", "session-token", cookie.Value)
		}
		csrfCookie := findSetCookie(t, w, constants.CSRFCookieName)
		if csrfCookie.Value != "csrf-token" {
			t.Fatalf("Expected CSRF cookie value %q, got %q", "csrf-token", csrfCookie.Value)
		}
		if csrfCookie.HttpOnly {
			t.Fatal("Expected CSRF cookie to be readable by frontend JavaScript")
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Expected valid JSON response: %v", err)
		}
		if _, exists := response["session_token"]; exists {
			t.Fatal("Expected session_token to be omitted from JSON response")
		}
		if _, exists := response["csrf_token"]; exists {
			t.Fatal("Expected csrf_token to be omitted from JSON response")
		}
	})

	t.Run("returns bad request for invalid json", func(t *testing.T) {
		service := &fakeAuthService{}
		router := authRouter(service)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{`))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
		if service.loginCalled {
			t.Fatal("Expected login service not to be called")
		}
	})

	t.Run("returns service error", func(t *testing.T) {
		service := &fakeAuthService{
			loginErr: errors.NewApiError(
				http.StatusUnauthorized,
				errors.BadRequestError("INVALID_CREDENTIALS"),
			),
		}
		router := authRouter(service)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"username":"tiago","password":"strong-password"}`))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})
}

func TestLogout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("revokes current session and clears cookie", func(t *testing.T) {
		service := &fakeAuthService{}
		router := authRouter(service)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
		req.AddCookie(&http.Cookie{Name: constants.SessionCookieName, Value: "session-token"})

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if !service.logoutCalled {
			t.Error("Expected logout service to be called")
		}
		if service.logoutToken != "session-token" {
			t.Errorf("Expected logout token %q, got %q", "session-token", service.logoutToken)
		}

		cookie := findSetCookie(t, w, constants.SessionCookieName)
		if cookie.MaxAge != -1 {
			t.Errorf("Expected clear cookie max age -1, got %d", cookie.MaxAge)
		}
		if cookie.Value != "" {
			t.Errorf("Expected cleared cookie value to be blank, got %q", cookie.Value)
		}
		if !cookie.HttpOnly {
			t.Error("Expected cleared cookie to be HttpOnly")
		}
		csrfCookie := findSetCookie(t, w, constants.CSRFCookieName)
		if csrfCookie.MaxAge != -1 {
			t.Errorf("Expected clear CSRF cookie max age -1, got %d", csrfCookie.MaxAge)
		}
		if csrfCookie.Value != "" {
			t.Errorf("Expected cleared CSRF cookie value to be blank, got %q", csrfCookie.Value)
		}
		if csrfCookie.HttpOnly {
			t.Error("Expected cleared CSRF cookie not to be HttpOnly")
		}
		assertSetCookieContains(t, w, "SameSite=Lax")
	})

	t.Run("returns unauthorized when logout service rejects missing cookie", func(t *testing.T) {
		service := &fakeAuthService{
			logoutErr: errors.NewApiError(
				http.StatusUnauthorized,
				errors.BadRequestError("INVALID_SESSION"),
			),
		}
		router := authRouter(service)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
		if !service.logoutCalled {
			t.Error("Expected logout service to be called")
		}
		if service.logoutToken != "" {
			t.Errorf("Expected blank logout token, got %q", service.logoutToken)
		}
	})
}

func TestSession(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns current authenticated session", func(t *testing.T) {
		expiresAt := time.Now().Add(time.Hour)
		service := &fakeAuthService{
			sessionResponse: dto.AuthSessionResponse{
				ID:        1,
				Name:      "Tiago",
				Username:  "tiago",
				Email:     "tiago@example.com",
				CPF:       "00000000000",
				ExpiresAt: expiresAt,
			},
		}
		router := authRouter(service)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/auth/session", nil)
		req.AddCookie(&http.Cookie{Name: constants.SessionCookieName, Value: "session-token"})

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if !service.sessionCalled {
			t.Fatal("Expected session service to be called")
		}
		if service.sessionToken != "session-token" {
			t.Fatalf("Expected session token %q, got %q", "session-token", service.sessionToken)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Expected valid JSON response: %v", err)
		}
		if response["username"] != "tiago" {
			t.Fatalf("Expected username %q, got %q", "tiago", response["username"])
		}
		if _, exists := response["session_token"]; exists {
			t.Fatal("Expected session_token to be omitted from JSON response")
		}
	})

	t.Run("returns unauthorized without session cookie", func(t *testing.T) {
		service := &fakeAuthService{}
		router := authRouter(service)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/auth/session", nil)

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
		if service.sessionCalled {
			t.Fatal("Expected session service not to be called")
		}
	})

	t.Run("returns service error for invalid session", func(t *testing.T) {
		service := &fakeAuthService{
			sessionErr: errors.NewApiError(
				http.StatusUnauthorized,
				errors.BadRequestError("INVALID_SESSION"),
			),
		}
		router := authRouter(service)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/auth/session", nil)
		req.AddCookie(&http.Cookie{Name: constants.SessionCookieName, Value: "bad-token"})

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})
}

func authRouter(service *fakeAuthService) *gin.Engine {
	router := gin.New()
	handler := NewAuthHandler(service)
	router.POST("/auth/login", handler.Login())
	router.GET("/auth/session", handler.Session())
	router.POST("/auth/logout", handler.Logout())
	return router
}

type fakeAuthService struct {
	loginCalled      bool
	loginErr         errors.ApiError
	loginRequest     dto.LoginRequest
	loginResponse    dto.LoginResponse
	loginSessionInfo dto.LoginSessionInfo
	logoutCalled     bool
	logoutErr        errors.ApiError
	logoutToken      string
	sessionCalled    bool
	sessionErr       errors.ApiError
	sessionResponse  dto.AuthSessionResponse
	sessionToken     string
}

func (f *fakeAuthService) Login(_ context.Context, request dto.LoginRequest, sessionInfo dto.LoginSessionInfo) (dto.LoginResponse, errors.ApiError) {
	f.loginCalled = true
	f.loginRequest = request
	f.loginSessionInfo = sessionInfo

	if f.loginErr != nil {
		return dto.LoginResponse{}, f.loginErr
	}

	return f.loginResponse, nil
}

func (f *fakeAuthService) Logout(_ context.Context, token string) errors.ApiError {
	f.logoutCalled = true
	f.logoutToken = token
	return f.logoutErr
}

func (f *fakeAuthService) Session(_ context.Context, token string) (dto.AuthSessionResponse, errors.ApiError) {
	f.sessionCalled = true
	f.sessionToken = token

	if f.sessionErr != nil {
		return dto.AuthSessionResponse{}, f.sessionErr
	}

	return f.sessionResponse, nil
}

func (f *fakeAuthService) ValidateSession(context.Context, string) (models.User, errors.ApiError) {
	return models.User{}, nil
}

func (f *fakeAuthService) ValidateCSRF(context.Context, string, string) errors.ApiError {
	return nil
}

func findSetCookie(t *testing.T, w *httptest.ResponseRecorder, name string) *http.Cookie {
	t.Helper()

	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}

	t.Fatalf("Expected cookie %q to be set", name)
	return nil
}

func assertSetCookieContains(t *testing.T, w *httptest.ResponseRecorder, want string) {
	t.Helper()

	for _, header := range w.Result().Header.Values("Set-Cookie") {
		if strings.Contains(header, want) {
			return
		}
	}

	t.Fatalf("Expected Set-Cookie header to contain %q, got %q", want, w.Result().Header.Values("Set-Cookie"))
}
