package dto

import (
	"strings"
	"time"

	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/go-commons/utils"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	CPF          string    `json:"cpf"`
	SessionToken string    `json:"-"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type LoginSessionInfo struct {
	UserAgent string
	IPAddress string
}

func (dto LoginRequest) Validate() []errors.ApiErrorItem {
	errs := make([]errors.ApiErrorItem, 0)

	if utils.IsBlank(dto.Username) || len(strings.TrimSpace(dto.Username)) <= 3 {
		errs = append(errs, errors.InvalidFieldError("USERNAME_INVALID_CHARS_COUNT"))
	}

	if utils.IsBlank(dto.Password) || len(strings.TrimSpace(dto.Password)) <= 6 {
		errs = append(errs, errors.InvalidFieldError("PASSWORD_INVALID_CHARS_COUNT"))
	}

	return errs
}
