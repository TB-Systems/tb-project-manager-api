package handlers

import (
	"net/http"

	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/go-commons/utils"
	"github.com/TB-Systems/tb-project-manager-api/constants"
	"github.com/TB-Systems/tb-project-manager-api/dto"
	"github.com/TB-Systems/tb-project-manager-api/services"
	"github.com/gin-gonic/gin"
)

type Auth struct {
	service services.Auth
}

func NewAuthHandler(service services.Auth) *Auth {
	return &Auth{service: service}
}

func (h *Auth) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		request, apiErr := utils.DecodeValidJson[dto.LoginRequest](c)
		if apiErr != nil {
			utils.SendErrorResponse(c, apiErr)
			return
		}

		sessionInfo := dto.LoginSessionInfo{
			UserAgent: c.GetHeader("User-Agent"),
			IPAddress: c.ClientIP(),
		}

		data, apiErr := h.service.Login(c.Request.Context(), request, sessionInfo)
		if apiErr != nil {
			utils.SendErrorResponse(c, apiErr)
			return
		}

		utils.SetSessionToken(c, constants.SessionCookieName, data.SessionToken, data.ExpiresAt)
		utils.SetCSRFToken(c, constants.CSRFCookieName, data.CSRFToken, data.ExpiresAt)

		utils.SendResponse(c, data, http.StatusOK)
	}
}

func invalidSession() errors.ApiError {
	return errors.NewApiError(
		http.StatusUnauthorized,
		errors.BadRequestError("INVALID_SESSION"),
	)
}

func (h *Auth) Session() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(constants.SessionCookieName)
		if err != nil {
			utils.SendErrorResponse(c, invalidSession())
			return
		}

		data, apiErr := h.service.Session(c.Request.Context(), token)
		if apiErr != nil {
			utils.SendErrorResponse(c, apiErr)
			return
		}

		utils.SendResponse(c, data, http.StatusOK)
	}
}

func (h *Auth) Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _ := c.Cookie(constants.SessionCookieName)

		if apiErr := h.service.Logout(c.Request.Context(), token); apiErr != nil {
			utils.SendErrorResponse(c, apiErr)
			return
		}

		utils.ClearSessionToken(c, constants.SessionCookieName)
		utils.ClearCSRFToken(c, constants.CSRFCookieName)
		utils.SendResponse(c, gin.H{"logged_out": true}, http.StatusOK)
	}
}
