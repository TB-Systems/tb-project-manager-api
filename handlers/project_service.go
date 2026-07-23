package handlers

import (
	"net/http"

	"github.com/TB-Systems/go-commons/commonsmodels"
	"github.com/TB-Systems/go-commons/utils"
	"github.com/TB-Systems/tb-project-manager-api/dto"
	"github.com/TB-Systems/tb-project-manager-api/services"
	"github.com/gin-gonic/gin"
)

type ProjectService struct {
	service services.ProjectService
}

func NewProjectServiceHandler(service services.ProjectService) *ProjectService {
	return &ProjectService{service: service}
}

func (h *ProjectService) List() gin.HandlerFunc {
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

		filter := dto.ProjectServiceListFilter{
			ProjectID: c.Query("project_id"),
		}

		data, apiErr := h.service.List(c.Request.Context(), params, filter)
		if apiErr != nil {
			utils.SendErrorResponse(c, apiErr)
			return
		}

		utils.SendResponse(c, data, http.StatusOK)
	}
}

func (h *ProjectService) FindByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		data, apiErr := h.service.FindByID(c.Request.Context(), c.Param("id"))
		if apiErr != nil {
			utils.SendErrorResponse(c, apiErr)
			return
		}

		utils.SendResponse(c, data, http.StatusOK)
	}
}

func (h *ProjectService) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		request, apiErr := utils.DecodeValidJson[dto.ProjectServiceCreateRequest](c)
		if apiErr != nil {
			utils.SendErrorResponse(c, apiErr)
			return
		}

		data, apiErr := h.service.Create(c.Request.Context(), request)
		if apiErr != nil {
			utils.SendErrorResponse(c, apiErr)
			return
		}

		utils.SendResponse(c, data, http.StatusCreated)
	}
}

func (h *ProjectService) Update() gin.HandlerFunc {
	return func(c *gin.Context) {
		request, apiErr := utils.DecodeValidJson[dto.ProjectServiceRequest](c)
		if apiErr != nil {
			utils.SendErrorResponse(c, apiErr)
			return
		}

		data, apiErr := h.service.Update(c.Request.Context(), c.Param("id"), request)
		if apiErr != nil {
			utils.SendErrorResponse(c, apiErr)
			return
		}

		utils.SendResponse(c, data, http.StatusOK)
	}
}

func (h *ProjectService) Delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		if apiErr := h.service.Delete(c.Request.Context(), c.Param("id")); apiErr != nil {
			utils.SendErrorResponse(c, apiErr)
			return
		}

		utils.SendResponse(c, commonsmodels.NewResponseSuccess(), http.StatusOK)
	}
}
