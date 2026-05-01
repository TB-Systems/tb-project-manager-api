package handlers

import (
	"net/http"

	"github.com/TB-Systems/go-commons/utils"
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

		data, apiErr := h.service.Login(c.Request.Context(), request)
		if apiErr != nil {
			utils.SendErrorResponse(c, apiErr)
			return
		}

		utils.SendResponse(c, data, http.StatusOK)
	}
}
