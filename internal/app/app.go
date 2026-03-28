package app

import (
	"context"
	"log/slog"

	resendAPI "github.com/rwrrioe/sso/internal/adapters/mail/resend"
	"github.com/rwrrioe/sso/internal/adapters/postgresql"
	redisstorage "github.com/rwrrioe/sso/internal/adapters/redis"
	"github.com/rwrrioe/sso/internal/app/grpc"
	"github.com/rwrrioe/sso/internal/usecase/auth"
	"github.com/rwrrioe/sso/internal/usecase/code"
)

type Config struct {
	APIKey      string
	GRPCPort    int
	PostgresDSN string

	Redis redisstorage.Config

	Resend resendAPI.Options

	Auth auth.Config
	Code code.Config
}

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	ctx context.Context,
	log *slog.Logger,
	cfg *Config,
) *App {
	postgresStorage, err := postgresql.New(ctx)
	if err != nil {
		panic(err)
	}

	redisStorage, err := redisstorage.New(log, &cfg.Redis)
	if err != nil {
		panic(err)
	}

	mailer := resendAPI.New(
		log,
		&cfg.Resend,
		cfg.APIKey,
	)

	authService := auth.New(
		log,
		postgresStorage,
		postgresStorage,
		postgresStorage,
		postgresStorage,
		redisStorage,
		&cfg.Auth,
	)

	codeService := code.New(
		log,
		redisStorage,
		mailer,
		postgresStorage,
		&cfg.Code,
	)

	grpcApp := grpcapp.New(
		log,
		cfg.GRPCPort,
		authService,
		codeService,
	)

	return &App{
		GRPCServer: grpcApp,
	}
}
