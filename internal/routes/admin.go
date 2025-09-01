package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/malakagl/go-template/internal/api/handlers"
	"github.com/malakagl/go-template/pkg/repositories"
	"github.com/malakagl/go-template/pkg/services"
	"gorm.io/gorm"
)

func AddAdminRoutes(r *chi.Mux, db *gorm.DB) {
	apiKeyRepo := repositories.NewApiKeyRepository(db)
	adminService := services.NewEndpointService(apiKeyRepo)
	apiKeyService := services.NewAPIKeyService(apiKeyRepo)
	adminHandler := handlers.NewAdminHandler(&adminService, &apiKeyService)

	r.Get("/admin/endpoints", adminHandler.GetEndpoints)
	r.Post("/admin/apikeys", adminHandler.CreateAPIKeys)
}
