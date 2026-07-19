package dto

import (
	"encoding/json"
	"time"

	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
)

type ServiceCheckRequest struct {
	ProjectServiceID uuid.UUID            `json:"project_service_id"`
	Status           models.ServiceStatus `json:"status"`
	StatusCode       int                  `json:"status_code"`
	ResponseTimeMS   int                  `json:"response_time_ms"`
	Message          json.RawMessage      `json:"message"`
}

type ServiceCheckResponse struct {
	ID               uuid.UUID            `json:"id"`
	ProjectServiceID uuid.UUID            `json:"project_service_id"`
	Status           models.ServiceStatus `json:"status"`
	StatusCode       int                  `json:"status_code"`
	ResponseTimeMS   int                  `json:"response_time_ms"`
	Message          json.RawMessage      `json:"message"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
}

func (request ServiceCheckRequest) Validate() []errors.ApiErrorItem {
	errs := make([]errors.ApiErrorItem, 0)

	if request.ProjectServiceID == uuid.Nil {
		errs = append(errs, errors.InvalidFieldError("SERVICE_CHECK_PROJECT_SERVICE_ID_INVALID"))
	}

	if !request.Status.IsValid() {
		errs = append(errs, errors.InvalidFieldError("SERVICE_CHECK_STATUS_INVALID"))
	}

	if request.StatusCode < 0 {
		errs = append(errs, errors.InvalidFieldError("SERVICE_CHECK_STATUS_CODE_INVALID"))
	}

	if request.ResponseTimeMS < 0 {
		errs = append(errs, errors.InvalidFieldError("SERVICE_CHECK_RESPONSE_TIME_INVALID"))
	}

	if len(request.Message) > 0 && !json.Valid(request.Message) {
		errs = append(errs, errors.InvalidFieldError("SERVICE_CHECK_MESSAGE_INVALID"))
	}

	return errs
}

func ServiceCheckResponseFromModel(serviceCheck models.ServiceCheck) ServiceCheckResponse {
	return ServiceCheckResponse{
		ID:               serviceCheck.ID,
		ProjectServiceID: serviceCheck.ProjectServiceID,
		Status:           serviceCheck.Status,
		StatusCode:       serviceCheck.StatusCode,
		ResponseTimeMS:   serviceCheck.ResponseTimeMS,
		Message:          serviceCheck.Message,
		CreatedAt:        serviceCheck.CreatedAt,
		UpdatedAt:        serviceCheck.UpdatedAt,
	}
}
