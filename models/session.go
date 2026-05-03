package models

import "time"

type UserSession struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     uint       `gorm:"uniqueIndex;not null" json:"user_id"`
	User       User       `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	TokenHash  string     `gorm:"size:64;uniqueIndex;not null" json:"-"`
	CSRFHash   string     `gorm:"size:64" json:"-"`
	UserAgent  string     `gorm:"size:512" json:"user_agent"`
	IPAddress  string     `gorm:"size:64" json:"ip_address"`
	ExpiresAt  time.Time  `gorm:"not null" json:"expires_at"`
	LastSeenAt time.Time  `gorm:"not null" json:"last_seen_at"`
	RevokedAt  *time.Time `json:"revoked_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
