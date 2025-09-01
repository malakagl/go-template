package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/malakagl/go-template/internal/api/handlers"
	"github.com/malakagl/go-template/pkg/repositories"
	"github.com/malakagl/go-template/pkg/services"
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
