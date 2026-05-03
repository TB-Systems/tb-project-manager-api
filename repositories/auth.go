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
	FindValidSessionByTokenHash(ctx context.Context, tokenHash string, now time.Time) (models.UserSession, error)
	RevokeSessionByTokenHash(ctx context.Context, tokenHash string, revokedAt time.Time) error
	TouchSession(ctx context.Context, sessionID uint, lastSeenAt time.Time) error
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

func (a auth) FindValidSessionByTokenHash(ctx context.Context, tokenHash string, now time.Time) (models.UserSession, error) {
	var session models.UserSession
	err := a.db.WithContext(ctx).
		Preload("User").
		Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", tokenHash, now).
		First(&session).
		Error
	if err != nil {
		return models.UserSession{}, err
	}

	return session, nil
}

func (a auth) RevokeSessionByTokenHash(ctx context.Context, tokenHash string, revokedAt time.Time) error {
	return a.db.WithContext(ctx).
		Model(&models.UserSession{}).
		Where("token_hash = ? AND revoked_at IS NULL", tokenHash).
		Updates(map[string]interface{}{
			"revoked_at": revokedAt,
			"updated_at": revokedAt,
		}).
		Error
}

func (a auth) TouchSession(ctx context.Context, sessionID uint, lastSeenAt time.Time) error {
	return a.db.WithContext(ctx).
		Model(&models.UserSession{}).
		Where("id = ?", sessionID).
		Updates(map[string]interface{}{
			"last_seen_at": lastSeenAt,
			"updated_at":   lastSeenAt,
		}).
		Error
}

func (a auth) UpsertSession(ctx context.Context, session models.UserSession) error {
	now := time.Now().UTC()

	return a.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"token_hash":   session.TokenHash,
			"csrf_hash":    session.CSRFHash,
			"user_agent":   session.UserAgent,
			"ip_address":   session.IPAddress,
			"expires_at":   session.ExpiresAt,
			"last_seen_at": session.LastSeenAt,
			"revoked_at":   nil,
			"updated_at":   now,
		}),
	}).Create(&session).Error
}
