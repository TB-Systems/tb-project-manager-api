package dto

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/go-commons/utils"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
)

type ServiceLogRequest struct {
	ProjectServiceID uuid.UUID       `json:"project_service_id"`
	Level            models.LogLevel `json:"level"`
	Event            string          `json:"event"`
	Message          json.RawMessage `json:"message"`
	Time             time.Time       `json:"time"`
}

type ServiceLogResponse struct {
	ID               uuid.UUID       `json:"id"`
	ProjectServiceID uuid.UUID       `json:"project_service_id"`
	Level            models.LogLevel `json:"level"`
	Event            string          `json:"event"`
	Message          json.RawMessage `json:"message"`
	Time             time.Time       `json:"time"`
	CreatedAt        time.Time       `json:"created_at"`
}

func (request ServiceLogRequest) Validate() []errors.ApiErrorItem {
	errs := make([]errors.ApiErrorItem, 0)

	if request.ProjectServiceID == uuid.Nil {
		errs = append(errs, errors.InvalidFieldError("SERVICE_LOG_PROJECT_SERVICE_ID_INVALID"))
	}

	if !request.Level.IsValid() {
		errs = append(errs, errors.InvalidFieldError("SERVICE_LOG_LEVEL_INVALID"))
	}

	if utils.IsBlank(request.Event) || len(strings.TrimSpace(request.Event)) > 100 {
		errs = append(errs, errors.InvalidFieldError("SERVICE_LOG_EVENT_INVALID"))
	}

	if len(request.Message) == 0 || !json.Valid(request.Message) {
		errs = append(errs, errors.InvalidFieldError("SERVICE_LOG_MESSAGE_INVALID"))
	}

	if request.Time.IsZero() {
		errs = append(errs, errors.InvalidFieldError("SERVICE_LOG_TIME_INVALID"))
	}

	return errs
}

func ServiceLogResponseFromModel(serviceLog models.ServiceLog) ServiceLogResponse {
	return ServiceLogResponse{
		ID:               serviceLog.ID,
		ProjectServiceID: serviceLog.ProjectServiceID,
		Level:            serviceLog.Level,
		Event:            serviceLog.Event,
		Message:          serviceLog.Message,
		Time:             serviceLog.Time,
		CreatedAt:        serviceLog.CreatedAt,
	}
}
