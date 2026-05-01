package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Project struct {
}

func NewProjectsHandler() *Project {
	return &Project{}
}

func (h *Project) List() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": []string{}})
	}
}

func (h *Project) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, "success")
	}
}
