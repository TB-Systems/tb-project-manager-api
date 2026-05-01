package handlers

import (
	"net/http"

	"github.com/TB-Systems/tb-project-manager-api/dto"
	"github.com/gin-gonic/gin"
)

type Auth struct {
}

func NewAuthHandler() *Auth {
	return &Auth{}
}

func (h *Auth) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request dto.LoginRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, "success")
	}
}
