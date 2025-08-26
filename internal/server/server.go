package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/malakagl/kart-challenge/internal/config"
	"github.com/malakagl/kart-challenge/internal/couponcode"
	"github.com/malakagl/kart-challenge/internal/database"
	"github.com/malakagl/kart-challenge/internal/middleware"
	"github.com/malakagl/kart-challenge/internal/routes"
	"github.com/malakagl/kart-challenge/pkg/errors"
	"github.com/malakagl/kart-challenge/pkg/log"
	"github.com/malakagl/kart-challenge/pkg/otel"
	"gorm.io/gorm"
)

type Server struct {
	httpServer *http.Server
	ErrChan    chan error
	db         *gorm.DB
	cfg        *config.Config
}

func NewServer(c *config.Config) *Server {
	return &Server{
		ErrChan: make(chan error, 1),
		cfg:     c,
	}
}

// Start sets up the database, coupon codes, routes, and starts the HTTP server.
// Returns the server instance for later shutdown.
func (s *Server) Start() error {
	ctx := context.Background()
	if s.cfg.Telemetry.Enabled {
		_, err := otel.InitTracer("kart-challenge", fmt.Sprintf("%s:%d", s.cfg.Telemetry.Host, s.cfg.Telemetry.Port))
		if err != nil {
			return err
		}
	}

	if err := database.RunMigrations(ctx, &s.cfg.Database); err != nil {
		log.Error().Err(err).Msgf("database migrations failed.")
		return err
	}

	couponcode.SetCouponCodeFiles(s.cfg.CouponCode.FilePaths)
	go func(ctx context.Context) {
		ctx, span := otel.Tracer(ctx, "decompress-coupon-files")
		defer span.End()

		log.Info().Msg("Started decompressing coupon code files in background")
		if errD := couponcode.SetupCouponCodeFiles(ctx, s.cfg.CouponCode.FilePaths); errD != nil {
			span.RecordError(errD)
			log.Error().Err(errD).Msg("coupon code setup failed.")
		}
	}(ctx)

	couponcode.InitCache(s.cfg.Server.MaxCouponCodeCacheSize)
	log.Info().Msgf("connecting to database")
	var err error
	s.db, err = database.Connect(ctx, &s.cfg.Database)
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to database.")
		return err
	}

	log.Info().Msgf("creating routes")
	middleware.SetRateLimits(s.cfg.Server.ReqLimitPerIP, s.cfg.Server.ReqBurstPerIP, s.cfg.Server.ReqRateWindow)
	r := chi.NewRouter()
	r.Use(middleware.Trace, middleware.Logging, middleware.Authentication, middleware.RateLimit)
	routes.AddHealthCheckRoutes(r)
	routes.AddProductRoutes(r, s.db)
	routes.AddOrderRoutes(r, s.db)

	serverAddr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	s.httpServer = &http.Server{
		Addr:    serverAddr,
		Handler: r,
	}

	go func() {
		if errSrv := s.httpServer.ListenAndServe(); errSrv != nil && !errors.Is(errSrv, http.ErrServerClosed) {
			log.Error().Err(errSrv).Msgf("HTTP server failed.")
			s.ErrChan <- errSrv
		}
	}()

	log.Info().Msgf("Host started on %s", serverAddr)
	return nil
}

// Stop gracefully shuts down the server with a context timeout.
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer != nil {
		log.Info().Msg("Shutting down server gracefully")
		if err := s.httpServer.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("Host shutdown error.")
		}
	}

	if s.db != nil {
		db, err := s.db.DB()
		if err == nil {
			if err = db.Close(); err != nil {
				log.Error().Err(err).Msg("error while closing sql.DB")
			} else {
				log.Info().Msg("DB closed successfully")
			}
		} else {
			log.Error().Err(err).Msg("failed to retrieve sql.DB for closing")
		}
	}

	log.Info().Msg("Host stopped successfully")
	return nil
}
