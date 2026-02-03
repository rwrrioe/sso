package main

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/rwrrioe/sso/internal/app"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	env := os.Getenv("LOGGER_ENV")
	portStr := os.Getenv("GRPC_PORT")
	if portStr == "" {
		portStr = "9081"
	}

	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 || port > 65535 {
		panic("invalid GRPC_PORT: " + portStr)
	}

	log := setupLogger(env)

	log.Info("starting app", slog.Any("env", env))
	application := app.New(ctx, log, port, time.Minute)

	if err := application.GRPCServer.Run(); err != nil {
		panic(err)
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
