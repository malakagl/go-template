package db

import "time"

type Endpoint struct {
	ID           uint      `gorm:"primaryKey"`
	HTTPMethod   string    `gorm:"not null"`
	HTTPEndpoint string    `gorm:"not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`

	// Associations
	APIKeys []APIKey `gorm:"many2many:api_key_endpoints;"`
}
