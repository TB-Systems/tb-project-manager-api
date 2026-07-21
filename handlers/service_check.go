package handlers

import (
	"net/http"

	"github.com/TB-Systems/go-commons/commonsmodels"
	"github.com/TB-Systems/go-commons/utils"
	"github.com/TB-Systems/tb-project-manager-api/services"
	"github.com/gin-gonic/gin"
)

type ServiceCheck struct {
	service services.ServiceCheck
}

func NewServiceCheckHandler(service services.ServiceCheck) *ServiceCheck {
	return &ServiceCheck{service: service}
}

func (h *ServiceCheck) List() gin.HandlerFunc {
	return func(c *gin.Context) {
		page, apiErr := utils.GetQueryPage(c)
		if apiErr != nil {
			utils.SendErrorResponse(c, apiErr)
			return
		}

		limit := utils.GetQueryLimit(c)
		params := commonsmodels.PaginatedParams{
			Limit:  limit,
			Offset: utils.CalculateOffset(page, limit),
			Page:   page,
		}

		data, apiErr := h.service.List(c.Request.Context(), params)
		if apiErr != nil {
			utils.SendErrorResponse(c, apiErr)
			return
		}

		utils.SendResponse(c, data, http.StatusOK)
	}
}

func (h *ServiceCheck) FindByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		data, apiErr := h.service.FindByID(c.Request.Context(), c.Param("id"))
		if apiErr != nil {
			utils.SendErrorResponse(c, apiErr)
			return
		}

		utils.SendResponse(c, data, http.StatusOK)
	}
}
