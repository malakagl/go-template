package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/malakagl/kart-challenge/pkg/models/dto/response"
)

func AddHealthCheckRoutes(r *chi.Mux) {
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, http.StatusOK, "ok")
	})
}
