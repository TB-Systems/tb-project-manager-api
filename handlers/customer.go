package handlers

import (
	"net/http"

	"github.com/TB-Systems/go-commons/commonsmodels"
	"github.com/TB-Systems/go-commons/utils"
	"github.com/TB-Systems/tb-project-manager-api/dto"
	"github.com/TB-Systems/tb-project-manager-api/services"
	"github.com/gin-gonic/gin"
)

type Customer struct {
	service services.Customer
}

func NewCustomerHandler(service services.Customer) *Customer {
	return &Customer{service: service}
}

func (h *Customer) List() gin.HandlerFunc {
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

func (h *Customer) FindByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		data, apiErr := h.service.FindByID(c.Request.Context(), c.Param("id"))
		if apiErr != nil {
			utils.SendErrorResponse(c, apiErr)
			return
		}

		utils.SendResponse(c, data, http.StatusOK)
	}
}

func (h *Customer) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		request, apiErr := utils.DecodeValidJson[dto.CustomerRequest](c)
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

func (h *Customer) Update() gin.HandlerFunc {
	return func(c *gin.Context) {
		request, apiErr := utils.DecodeValidJson[dto.CustomerRequest](c)
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

func (h *Customer) Delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		if apiErr := h.service.Delete(c.Request.Context(), c.Param("id")); apiErr != nil {
			utils.SendErrorResponse(c, apiErr)
			return
		}

		utils.SendResponse(c, commonsmodels.NewResponseSuccess(), http.StatusOK)
	}
}
