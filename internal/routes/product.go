package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/malakagl/go-template/internal/api/handlers"
	"github.com/malakagl/go-template/pkg/repositories"
	"github.com/malakagl/go-template/pkg/services"
	"gorm.io/gorm"
)

func AddProductRoutes(r *chi.Mux, db *gorm.DB) {
	productRepo := repositories.NewProductRepo(db)
	productService := services.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(&productService)
	r.Get("/products", productHandler.ListProducts)
	r.Get("/products/{productID}", productHandler.GetProductByID)
}
