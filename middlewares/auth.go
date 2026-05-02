package middlewares

import (
	"net/http"

	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/go-commons/utils"
	"github.com/TB-Systems/tb-project-manager-api/constants"
	"github.com/TB-Systems/tb-project-manager-api/services"
	"github.com/gin-gonic/gin"
)

func AuthRequired(authService services.Auth) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(constants.SessionCookieName)
		if err != nil {
			utils.SendErrorResponse(c, invalidSession())
			return
		}

		user, apiErr := authService.ValidateSession(c.Request.Context(), token)
		if apiErr != nil {
			utils.SendErrorResponse(c, apiErr)
			return
		}

		c.Set(constants.AuthUserIDKey, user.ID)
		c.Set(constants.AuthUserKey, user)
		c.Next()
	}
}

func invalidSession() errors.ApiError {
	return errors.NewApiError(
		http.StatusUnauthorized,
		errors.BadRequestError("INVALID_SESSION"),
	)
}
