package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/tb-project-manager-api/constants"
	"github.com/TB-Systems/tb-project-manager-api/dto"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/gin-gonic/gin"
)

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

func authRouter(service *fakeAuthService) *gin.Engine {
	router := gin.New()
	handler := NewAuthHandler(service)
	router.POST("/auth/logout", handler.Logout())
	return router
}

type fakeAuthService struct {
	logoutCalled bool
	logoutErr    errors.ApiError
	logoutToken  string
}

func (f *fakeAuthService) Login(context.Context, dto.LoginRequest, dto.LoginSessionInfo) (dto.LoginResponse, errors.ApiError) {
	return dto.LoginResponse{}, nil
}

func (f *fakeAuthService) Logout(_ context.Context, token string) errors.ApiError {
	f.logoutCalled = true
	f.logoutToken = token
	return f.logoutErr
}

func (f *fakeAuthService) ValidateSession(context.Context, string) (models.User, errors.ApiError) {
	return models.User{}, nil
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
