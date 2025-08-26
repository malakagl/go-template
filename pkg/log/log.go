package log

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/malakagl/kart-challenge/internal/config"
	"github.com/malakagl/kart-challenge/pkg/log/hooks"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger instance
var Logger zerolog.Logger

// Init initializes global logger
func Init(serviceName string, cfg config.LoggingConfig) {
	zerolog.TimeFieldFormat = time.RFC3339

	levelStr := strings.ToLower(cfg.Level)
	level, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		level = zerolog.InfoLevel
	}

	var output zerolog.Logger
	if cfg.JsonFormat {
		output = zerolog.New(os.Stderr)
	} else {
		output = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		})
	}

	Logger = output.Level(level).
		With().
		Timestamp().
		Str("service", serviceName).
		Logger()

	log.Logger = Logger
}

// WithCtx returns a logger with traceHook attached
func WithCtx(ctx context.Context) *zerolog.Logger {
	l := Logger.Hook(hooks.NewTraceHook(ctx))
	return &l
}

func Info() *zerolog.Event  { return Logger.Info() }
func Debug() *zerolog.Event { return Logger.Debug() }
func Warn() *zerolog.Event  { return Logger.Warn() }
func Error() *zerolog.Event { return Logger.Error() }
func Fatal() *zerolog.Event { return Logger.Fatal() }
