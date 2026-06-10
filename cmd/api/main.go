// Package main provides the entry point for the API server.
package main

import (
	"log/slog"
	"os"

	"github.com/golang-api-server/internal/config"
	"github.com/golang-api-server/internal/logger"
	"github.com/golang-api-server/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	// Phase 1: stdout-only logger for startup errors before Kafka is ready.
	slog.SetDefault(slog.New(newStdoutHandler(cfg)))

	deps, err := server.NewDependencies(cfg)
	if err != nil {
		slog.Error("init dependencies", "error", err)
		os.Exit(1)
	}
	defer deps.Close()

	// Phase 2: tee to both stdout and Kafka now that the producer is ready.
	kafkaH := logger.NewKafkaHandler(deps.LogProducer, parseLevel(cfg.LogLevel))
	deps.LogHandler = kafkaH
	slog.SetDefault(slog.New(logger.NewTeeHandler(newStdoutHandler(cfg), kafkaH)))

	app := server.New(deps)
	if err := app.Run(); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

func newStdoutHandler(cfg *config.Config) slog.Handler {
	opts := &slog.HandlerOptions{Level: parseLevel(cfg.LogLevel)}
	if cfg.LogFormat == "text" {
		return slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.NewJSONHandler(os.Stdout, opts)
}

func parseLevel(s string) slog.Level {
	var level slog.Level
	_ = level.UnmarshalText([]byte(s))
	return level
}
