package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malakagl/kart-challenge/pkg/errors"
	"github.com/malakagl/kart-challenge/pkg/models/dto/request"
	"github.com/malakagl/kart-challenge/pkg/models/dto/response"
	"github.com/stretchr/testify/mock"
)

// MockOrderService implements OrderService for testing
type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) Create(_ context.Context, _ *request.OrderRequest) (*response.OrderResponse, error) {
	args := m.Called()
	return args.Get(0).(*response.OrderResponse), args.Error(1)
}

func TestCreateOrder(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		mockRes        *response.OrderResponse
		mockErr        error
		expectedStatus int
	}{
		{
			name: "successful order",
			body: request.OrderRequest{
				CouponCode: "HAPPYHRS",
				Items:      []request.Item{{ProductID: "1", Quantity: 2}},
			},
			mockRes:        &response.OrderResponse{ID: "1234", Total: 100.0},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid JSON",
			body:           "{invalid-json}",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid coupon code",
			body: request.OrderRequest{
				CouponCode: "WRONGCODE",
				Items:      []request.Item{{ProductID: "1", Quantity: 2}},
			},
			mockErr:        errors.ErrInvalidCouponCode,
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "invalid item count",
			body: request.OrderRequest{
				CouponCode: "TESTCODE",
				Items:      []request.Item{{ProductID: "1", Quantity: 0}},
			},
			mockErr:        errors.ErrInvalidCouponCode,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			body: request.OrderRequest{
				CouponCode: "HAPPYHRS",
				Items:      []request.Item{{ProductID: "1", Quantity: 2}},
			},
			mockErr:        errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyBytes []byte
			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			mockService := new(MockOrderService)
			handler := NewOrderHandler(mockService)
			mockService.On("Create", mock.Anything).Return(tt.mockRes, tt.mockErr)
			handler.CreateOrder(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}
