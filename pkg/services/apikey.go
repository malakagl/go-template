package services

import (
	"context"

	"github.com/malakagl/go-template/pkg/errors"
	"github.com/malakagl/go-template/pkg/models/dto/request"
	"github.com/malakagl/go-template/pkg/models/dto/response"
	"github.com/malakagl/go-template/pkg/repositories"
	"github.com/malakagl/go-template/pkg/util"
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
