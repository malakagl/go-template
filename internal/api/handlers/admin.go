package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/malakagl/go-template/pkg/errors"
	"github.com/malakagl/go-template/pkg/log"
	"github.com/malakagl/go-template/pkg/models/dto/request"
	"github.com/malakagl/go-template/pkg/models/dto/response"
	"github.com/malakagl/go-template/pkg/services"
	"github.com/malakagl/go-template/pkg/util"
)

type AdminHandler struct {
	endpointService services.IEndpointService
	apiKeyService   services.IAPIKeyService
}

func NewAdminHandler(o services.IEndpointService, a services.IAPIKeyService) *AdminHandler {
	return &AdminHandler{endpointService: o, apiKeyService: a}
}

func (a *AdminHandler) GetEndpoints(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	endpoints, err := a.endpointService.GetEndpoints(ctx)
	if err != nil && !errors.Is(err, errors.ErrEndpointsNotFound) {
		log.WithCtx(ctx).Error().Msgf("Error fetching endpoints: %v", err)
		response.Error(w, http.StatusInternalServerError, "Error fetching endpoints", "Error fetching endpoints")
		return
	}

	response.Success(w, http.StatusOK, endpoints)
}

func (a *AdminHandler) CreateAPIKeys(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var apiKeyReq request.ApiKeyRequest
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&apiKeyReq); err != nil {
		log.WithCtx(ctx).Error().Msgf("Error decoding request body: %v", err)
		response.Error(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	orderRes, err := a.apiKeyService.Create(ctx, &apiKeyReq)
	if err != nil {
		log.WithCtx(ctx).Error().Msgf("Error creating order: %v", err)
		code, msg := util.MapErrorToHTTP(err)
		response.Error(w, code, msg, err.Error())
		return
	}

	response.Success(w, http.StatusCreated, orderRes)
}
