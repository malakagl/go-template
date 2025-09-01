package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/malakagl/kart-challenge/internal/api/handlers"
	"github.com/malakagl/kart-challenge/pkg/repositories"
	"github.com/malakagl/kart-challenge/pkg/services"
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
