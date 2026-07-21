package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/tb-project-manager-api/constants"
	"github.com/TB-Systems/tb-project-manager-api/dto"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/gin-gonic/gin"
)

func TestAuthRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("rejects request without session cookie", func(t *testing.T) {
		authService := &fakeAuthService{}
		router := authRouter(authService)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
		if authService.validateCalled {
			t.Error("Expected auth service not to be called without cookie")
		}
	})

	t.Run("rejects invalid session", func(t *testing.T) {
		authService := &fakeAuthService{validateErr: invalidSession()}
		router := authRouter(authService)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.AddCookie(&http.Cookie{Name: constants.SessionCookieName, Value: "bad-token"})

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
		if !authService.validateCalled {
			t.Error("Expected auth service to be called")
		}
		if authService.receivedToken != "bad-token" {
			t.Errorf("Expected token %q, got %q", "bad-token", authService.receivedToken)
		}
	})

	t.Run("accepts valid session and sets auth context", func(t *testing.T) {
		authService := &fakeAuthService{user: models.User{ID: 42, Username: "tiago"}}
		router := authRouter(authService)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.AddCookie(&http.Cookie{Name: constants.SessionCookieName, Value: "valid-token"})

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if w.Body.String() != "42" {
			t.Errorf("Expected user ID body %q, got %q", "42", w.Body.String())
		}
		if !authService.validateCalled {
			t.Error("Expected auth service to be called")
		}
		if authService.receivedToken != "valid-token" {
			t.Errorf("Expected token %q, got %q", "valid-token", authService.receivedToken)
		}
	})
}

func authRouter(authService *fakeAuthService) *gin.Engine {
	router := gin.New()
	router.GET("/protected", AuthRequired(authService), func(c *gin.Context) {
		userID, ok := c.Get(constants.AuthUserIDKey)
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}

		c.String(http.StatusOK, "%d", userID)
	})

	return router
}

type fakeAuthService struct {
	user           models.User
	validateErr    errors.ApiError
	validateCalled bool
	receivedToken  string
}

func (f *fakeAuthService) Login(context.Context, dto.LoginRequest, dto.LoginSessionInfo) (dto.LoginResponse, errors.ApiError) {
	return dto.LoginResponse{}, nil
}

func (f *fakeAuthService) Logout(context.Context, string) errors.ApiError {
	return nil
}

func (f *fakeAuthService) Session(context.Context, string) (dto.AuthSessionResponse, errors.ApiError) {
	return dto.AuthSessionResponse{}, nil
}

func (f *fakeAuthService) ValidateSession(_ context.Context, token string) (models.User, errors.ApiError) {
	f.validateCalled = true
	f.receivedToken = token

	if f.validateErr != nil {
		return models.User{}, f.validateErr
	}

	return f.user, nil
}

func (f *fakeAuthService) ValidateCSRF(context.Context, string, string) errors.ApiError {
	return nil
}
