package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/malakagl/kart-challenge/pkg/log"
	"github.com/malakagl/kart-challenge/pkg/models/dto/request"
	"github.com/malakagl/kart-challenge/pkg/models/dto/response"
	"github.com/malakagl/kart-challenge/pkg/services"
	"github.com/malakagl/kart-challenge/pkg/util"
)

type OrderHandler struct {
	orderService services.IOrderService
	validator    *validator.Validate
}

func NewOrderHandler(o services.IOrderService) *OrderHandler {
	return &OrderHandler{
		orderService: o,
		validator:    validator.New(),
	}
}

func (o *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var orderReq request.OrderRequest
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&orderReq); err != nil {
		log.WithCtx(ctx).Error().Msgf("Error decoding request body: %v", err)
		response.Error(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := o.validator.Struct(orderReq); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			errs := make([]string, len(ve))
			for i, fe := range ve {
				errs[i] = fmt.Sprintf("%s failed on %s", fe.Field(), fe.Tag())
			}
			response.Error(w, http.StatusBadRequest, "Invalid request data", strings.Join(errs, ", "))
			return
		}

		response.Error(w, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	orderRes, err := o.orderService.Create(ctx, &orderReq)
	if err != nil {
		log.WithCtx(ctx).Error().Msgf("Error creating order: %v", err)
		code, msg := util.MapErrorToHTTP(err)
		response.Error(w, code, msg, err.Error())
		return
	}

	response.Success(w, http.StatusCreated, orderRes)
}
