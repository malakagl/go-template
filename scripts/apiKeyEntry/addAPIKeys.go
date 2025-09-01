package main

import (
	"context"
	"flag"
	"strings"

	"github.com/malakagl/kart-challenge/internal/config"
	"github.com/malakagl/kart-challenge/internal/database"
	"github.com/malakagl/kart-challenge/pkg/log"
	"github.com/malakagl/kart-challenge/pkg/repositories"
)

func main() {
	ctx := context.Background()
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "config.yaml", "path to YAML config file")
	flag.Parse()

	log.Init("kart-challenge", config.LoggingConfig{Level: "info", JsonFormat: false})
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	db, err := database.Connect(ctx, &cfg.Database)
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to database.")
		return
	}

	apiKeyRepo := repositories.NewApiKeyRepository(db)
	endpoints, err := apiKeyRepo.FindAllEndpoints(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to find all endpoints")
		return
	}

	var adminEndPoints []uint
	for _, endpoint := range endpoints {
		if strings.HasPrefix(endpoint.HTTPEndpoint, "/admin/") {
			adminEndPoints = append(adminEndPoints, endpoint.ID)
		}
	}

	if len(adminEndPoints) != 0 {
		adminKey, err := apiKeyRepo.CreateAPIKeyWithEndpoints(ctx, adminEndPoints)
		if err != nil {
			log.Error().Err(err).Msg("failed to create admin key")
			return
		}

		log.Info().Msgf("created admin key: %s", adminKey)
		return
	}

	log.Info().Msg("no admin apis available with prefix '/admin/'")
}
