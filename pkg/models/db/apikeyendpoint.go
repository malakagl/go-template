package db

import "time"

type APIKeyEndpoint struct {
	APIKeyID   uint      `gorm:"primaryKey"`
	EndpointID uint      `gorm:"primaryKey"`
	IsActive   bool      `gorm:"default:true"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`

	// Associations
	APIKey   APIKey   `gorm:"foreignKey:APIKeyID"`
	Endpoint Endpoint `gorm:"foreignKey:EndpointID"`
}
