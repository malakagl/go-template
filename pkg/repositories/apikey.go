package repositories

import (
	"context"

	"github.com/malakagl/go-template/pkg/models/db"
	"github.com/malakagl/go-template/pkg/util"
	"gorm.io/gorm"
)

type ApiKeyRepository struct {
	db *gorm.DB
}

func NewApiKeyRepository(db *gorm.DB) *ApiKeyRepository {
	return &ApiKeyRepository{db: db}
}

func (a *ApiKeyRepository) GetEndPoints(clientID string) (*db.APIKey, error) {
	var key db.APIKey
	err := a.db.Preload("Endpoints").Where("client_id = ?", clientID).First(&key).Error
	return &key, err
}

// CreateAPIKeyWithEndpoints generates a key and associates it with endpoints
func (r *ApiKeyRepository) CreateAPIKeyWithEndpoints(ctx context.Context, endpointIDs []uint) (string, error) {
	clientID, secretHash, fullKey, err := util.GenerateAPIKey()
	if err != nil {
		return "", err
	}

	apiKey := db.APIKey{
		ClientID: clientID,
		APIKey:   secretHash,
	}

	err = r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(&apiKey).Error; err != nil {
			return err
		}

		for _, eid := range endpointIDs {
			link := db.APIKeyEndpoint{
				APIKeyID:   apiKey.ID,
				EndpointID: eid,
				IsActive:   true,
			}
			if err := tx.WithContext(ctx).Create(&link).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return fullKey, nil
}
