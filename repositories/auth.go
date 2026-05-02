package repositories

import (
	"context"
	"time"

	"github.com/TB-Systems/tb-project-manager-api/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Auth interface {
	Authenticate(ctx context.Context, login models.Login) (models.User, error)
	UpsertSession(ctx context.Context, session models.UserSession) error
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

func (a auth) UpsertSession(ctx context.Context, session models.UserSession) error {
	now := time.Now().UTC()

	return a.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"token_hash":   session.TokenHash,
			"user_agent":   session.UserAgent,
			"ip_address":   session.IPAddress,
			"expires_at":   session.ExpiresAt,
			"last_seen_at": session.LastSeenAt,
			"revoked_at":   nil,
			"updated_at":   now,
		}),
	}).Create(&session).Error
}
