package services

import (
	"context"
	"net/http"

	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/tb-project-manager-api/dto"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/TB-Systems/tb-project-manager-api/repositories"
)

type Auth interface {
	Login(ctx context.Context, request dto.LoginRequest) (dto.LoginResponse, errors.ApiError)
}

type auth struct {
	repository repositories.Auth
}

func NewAuthService(repository repositories.Auth) Auth {
	return auth{repository: repository}
}

func (a auth) Login(ctx context.Context, request dto.LoginRequest) (dto.LoginResponse, errors.ApiError) {
	login := models.Login{
		Username: request.Username,
		Password: request.Password,
	}

	user, err := a.repository.Authenticate(ctx, login)
	if err != nil {
		return dto.LoginResponse{}, errors.NewApiError(
			http.StatusUnauthorized,
			errors.BadRequestError("INVALID_CREDENTIALS"),
		)
	}

	return dto.LoginResponse{
		ID:       user.ID,
		Name:     user.Name,
		Username: user.Username,
		Email:    user.Email,
		CPF:      user.CPF,
	}, nil
}
