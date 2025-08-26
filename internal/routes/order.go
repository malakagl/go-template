package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/malakagl/kart-challenge/internal/api/handlers"
	"github.com/malakagl/kart-challenge/pkg/repositories"
	"github.com/malakagl/kart-challenge/pkg/services"
	"gorm.io/gorm"
)

func AddOrderRoutes(r *chi.Mux, db *gorm.DB) {
	productRepo := repositories.NewProductRepo(db)
	orderRepo := repositories.NewOrderRepo(db)
	couponCodeRepo := repositories.NewCouponCodeRepository(db)
	orderService := services.NewOrderService(orderRepo, couponCodeRepo, productRepo)
	orderHandler := handlers.NewOrderHandler(&orderService)

	r.Post("/orders", orderHandler.CreateOrder)
}
