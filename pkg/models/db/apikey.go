package db

import "time"

type APIKey struct {
	ID        uint      `gorm:"primaryKey"`
	ClientID  string    `gorm:"not null"`
	APIKey    string    `gorm:"not null"` // store bcrypt hash of the secret
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	// Associations
	Endpoints []Endpoint `gorm:"many2many:api_key_endpoints;"`
}
