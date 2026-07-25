package handlers

import (
	"net/http"

	"github.com/TB-Systems/go-commons/utils"
	"github.com/TB-Systems/tb-project-manager-api/dto"
	"github.com/TB-Systems/tb-project-manager-api/services"
	"github.com/gin-gonic/gin"
)

type CustomerProjectInvoice struct {
	service services.CustomerProjectInvoice
}

func NewCustomerProjectInvoiceHandler(service services.CustomerProjectInvoice) *CustomerProjectInvoice {
	return &CustomerProjectInvoice{service: service}
}

func (h *CustomerProjectInvoice) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		request, apiErr := utils.DecodeValidJson[dto.CustomerProjectInvoiceRequest](c)
		if apiErr != nil {
			utils.SendErrorResponse(c, apiErr)
			return
		}

		data, apiErr := h.service.Create(c.Request.Context(), c.Param("id"), request)
		if apiErr != nil {
			utils.SendErrorResponse(c, apiErr)
			return
		}

		utils.SendResponse(c, data, http.StatusCreated)
	}
}

func (h *CustomerProjectInvoice) Update() gin.HandlerFunc {
	return func(c *gin.Context) {
		request, apiErr := utils.DecodeValidJson[dto.CustomerProjectInvoiceRequest](c)
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

func (h *CustomerProjectInvoice) Pay() gin.HandlerFunc {
	return func(c *gin.Context) {
		request := dto.CustomerProjectInvoicePayRequest{}
		if c.Request.ContentLength != 0 {
			decodedRequest, apiErr := utils.DecodeValidJson[dto.CustomerProjectInvoicePayRequest](c)
			if apiErr != nil {
				utils.SendErrorResponse(c, apiErr)
				return
			}
			request = decodedRequest
		}

		data, apiErr := h.service.Pay(c.Request.Context(), c.Param("id"), request)
		if apiErr != nil {
			utils.SendErrorResponse(c, apiErr)
			return
		}

		utils.SendResponse(c, data, http.StatusOK)
	}
}

func (h *CustomerProjectInvoice) Unpay() gin.HandlerFunc {
	return func(c *gin.Context) {
		data, apiErr := h.service.Unpay(c.Request.Context(), c.Param("id"))
		if apiErr != nil {
			utils.SendErrorResponse(c, apiErr)
			return
		}

		utils.SendResponse(c, data, http.StatusOK)
	}
}
