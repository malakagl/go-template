package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/malakagl/kart-challenge/internal/api/handlers"
	"github.com/malakagl/kart-challenge/pkg/repositories"
	"github.com/malakagl/kart-challenge/pkg/services"
	"gorm.io/gorm"
)

func AddProductRoutes(r *chi.Mux, db *gorm.DB) {
	productRepo := repositories.NewProductRepo(db)
	productService := services.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(&productService)
	r.Get("/products", productHandler.ListProducts)
	r.Get("/products/{productID}", productHandler.GetProductByID)
}
