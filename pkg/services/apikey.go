package services

import (
	"context"

	"github.com/malakagl/kart-challenge/pkg/errors"
	"github.com/malakagl/kart-challenge/pkg/models/dto/request"
	"github.com/malakagl/kart-challenge/pkg/models/dto/response"
	"github.com/malakagl/kart-challenge/pkg/repositories"
	"github.com/malakagl/kart-challenge/pkg/util"
)

type IAPIKeyService interface {
	Create(ctx context.Context, req *request.ApiKeyRequest) (*response.APIKeyResponse, error)
}

type APIKeyService struct {
	apiKeyRepo *repositories.ApiKeyRepository
}

func NewAPIKeyService(a *repositories.ApiKeyRepository) APIKeyService {
	return APIKeyService{apiKeyRepo: a}
}

func (a *APIKeyService) Create(ctx context.Context, req *request.ApiKeyRequest) (*response.APIKeyResponse, error) {
	eps := make([]uint, len(req.EndPoints))
	for i := range req.EndPoints {
		var err error
		if eps[i], err = util.StringToUint(req.EndPoints[i]); err != nil {
			return &response.APIKeyResponse{}, errors.ErrBadRequest
		}
	}

	apiKey, err := a.apiKeyRepo.CreateAPIKeyWithEndpoints(ctx, eps)
	if err != nil {
		return nil, errors.ErrInternalServerError
	}

	return &response.APIKeyResponse{ApiKey: apiKey}, nil
}
