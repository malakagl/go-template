package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/malakagl/kart-challenge/pkg/errors"
	"github.com/malakagl/kart-challenge/pkg/models/dto/response"
	"github.com/stretchr/testify/mock"
)

// MockProductService implements ProductService for testing
type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) FindAll(_ context.Context) (*response.ProductsResponse, error) {
	args := m.Called()
	return args.Get(0).(*response.ProductsResponse), args.Error(1)
}

func (m *MockProductService) FindByID(_ context.Context, id uint) (*response.ProductResponse, error) {
	args := m.Called(id)
	return args.Get(0).(*response.ProductResponse), args.Error(1)
}

func TestListProducts(t *testing.T) {
	tests := []struct {
		name           string
		mockRes        *response.ProductsResponse
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "successful request",
			mockRes:        &response.ProductsResponse{Products: []response.Product{{ID: "1"}}},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "error response",
			mockErr:        errors.New("some error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "products not found",
			mockRes:        &response.ProductsResponse{Products: []response.Product{}},
			expectedStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/products", nil)
			w := httptest.NewRecorder()

			mockService := new(MockProductService)
			handler := NewProductHandler(mockService)
			mockService.On("FindAll").Return(tt.mockRes, tt.mockErr)
			handler.ListProducts(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestGetProductByID(t *testing.T) {
	tests := []struct {
		name           string
		id             uint
		mockRes        *response.ProductResponse
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "successful request",
			id:             uint(1),
			mockRes:        &response.ProductResponse{ID: "1"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "request with invalid product id",
			id:             uint(0),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "error finding product",
			id:             uint(1),
			mockErr:        errors.New("some error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "product not found",
			id:             uint(1),
			mockErr:        errors.ErrProductNotFound,
			expectedStatus: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := chi.NewRouteContext()
			ctx.URLParams.Add("productID", strconv.Itoa(int(tt.id)))
			req := httptest.NewRequest(http.MethodGet, "/products/"+strconv.Itoa(int(tt.id)), nil)
			req = req.WithContext(context.WithValue(context.Background(), chi.RouteCtxKey, ctx))
			w := httptest.NewRecorder()

			mockService := new(MockProductService)
			mockService.On("FindByID", tt.id).Return(tt.mockRes, tt.mockErr)

			handler := NewProductHandler(mockService)
			handler.GetProductByID(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}
