package services

import (
	"context"
	"strconv"

	"github.com/malakagl/kart-challenge/pkg/errors"
	"github.com/malakagl/kart-challenge/pkg/log"
	"github.com/malakagl/kart-challenge/pkg/models/dto/response"
	"github.com/malakagl/kart-challenge/pkg/repositories"
)

type IEndpointService interface {
	GetEndpoints(ctx context.Context) (*response.EndpointsResponse, error)
}

type EndpointService struct {
	apiKeyRepo *repositories.ApiKeyRepository
}

func NewEndpointService(a *repositories.ApiKeyRepository) EndpointService {
	return EndpointService{apiKeyRepo: a}
}

func (a *EndpointService) GetEndpoints(ctx context.Context) (*response.EndpointsResponse, error) {
	res, err := a.apiKeyRepo.FindAllEndpoints(ctx)
	if err != nil {
		log.WithCtx(ctx).Error().Msgf("findAll failed with error: %v", err)
		return nil, errors.ErrEndpointsNotFound
	}

	if res == nil {
		log.WithCtx(ctx).Error().Msg("findAll returned 0 elements")
		return nil, errors.ErrEndpointsNotFound
	}

	endpoints := make([]response.Endpoint, len(res))
	for i, e := range res {
		endpoints[i] = response.Endpoint{
			ID:           strconv.Itoa(int(e.ID)),
			HttpMethod:   string(e.HTTPMethod),
			HttpEndpoint: e.HTTPEndpoint,
		}
	}

	return &response.EndpointsResponse{Endpoints: endpoints}, nil
}
