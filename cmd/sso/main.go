package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/rwrrioe/sso/internal/adapters/mail/resend"
	redisstorage "github.com/rwrrioe/sso/internal/adapters/redis"
	"github.com/rwrrioe/sso/internal/app"
	"github.com/rwrrioe/sso/internal/config"
	"github.com/rwrrioe/sso/internal/usecase/auth"
	"github.com/rwrrioe/sso/internal/usecase/code"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	cfg := config.MustLoad()
	log := setupLogger(cfg.LoggerType)

	log.Info("starting app", slog.Any("port", cfg.GRPCPort))
	appCfg := setupAppConfig(cfg)
	application := app.New(ctx, log, appCfg)

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

func setupAppConfig(cfg *config.Config) *app.Config {
	return &app.Config{
		GRPCPort: cfg.GRPCPort,

		Redis: &redisstorage.Config{
			Address:  cfg.Redis.Address,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
			Protocol: cfg.Redis.Protocol,
		},

		Resend: &resend.Config{
			From: cfg.Resend.From,
			Name: cfg.Resend.Name,
		},

		Auth: &auth.Config{
			AccessTokenTTL:  cfg.Auth.AccessTokenTTL,
			RefreshTokenTTL: cfg.Auth.RefreshTokenTTL,
			ResetTokenTTL:   cfg.Auth.ResetTokenTTL,
		},

		Code: &code.Config{
			CodeTTL: cfg.Code.CodeTTL,
		},
	}

}
