package repositories

import (
	"context"

	"github.com/TB-Systems/tb-project-manager-api/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Auth interface {
	Authenticate(ctx context.Context, login models.Login) (models.User, error)
}

type auth struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) Auth {
	return auth{db: db}
}

func (a auth) Authenticate(ctx context.Context, login models.Login) (models.User, error) {
	var user models.User
	if err := a.db.WithContext(ctx).Where("username = ?", login.Username).First(&user).Error; err != nil {
		return models.User{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password)); err != nil {
		return models.User{}, err
	}

	return user, nil
}
