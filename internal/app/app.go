package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/rwrrioe/sso/internal/adapters/storage/postgresql"
	"github.com/rwrrioe/sso/internal/app/grpc"
	"github.com/rwrrioe/sso/internal/usecase/auth"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	ctx context.Context,
	log *slog.Logger,
	grpcPort int,
	tokenTTL time.Duration,
) *App {
	storage, err := postgresql.New(ctx)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, grpcPort, authService)

	return &App{
		GRPCServer: grpcApp,
	}
}
