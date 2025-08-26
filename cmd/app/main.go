package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/malakagl/kart-challenge/internal/config"
	"github.com/malakagl/kart-challenge/internal/server"
	"github.com/malakagl/kart-challenge/pkg/log"
)

func main() {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "config.yaml", "path to YAML config file")
	flag.Parse()

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	log.Init("kart-challenge", cfg.Logging)
	log.Info().Msgf("Host setting up on port: %d", cfg.Server.Port)

	s := server.NewServer(cfg)
	if err := s.Start(); err != nil {
		log.Fatal().Err(err).Msg("server failed to start")
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Wait for termination signal
	select {
	case sig := <-sigChan:
		log.Info().Msgf("Received signal: %s, shutting down...", sig)
	case srvErr := <-s.ErrChan:
		log.Error().Err(srvErr).Msg("Received error from server. shutting down...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.GracefulTimeout)
	defer cancel()

	if err := s.Stop(ctx); err != nil {
		log.Fatal().Err(err).Msg("server shutdown failed")
	}

	log.Info().Msg("Host exited gracefully")
}
