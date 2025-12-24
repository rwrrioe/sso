package app

import (
	"log/slog"

	grpcapp "github.com/rwrrioe/sso/internal/app/grpc"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
) *App {
	//todo init storage
	//todo init authservice

	grpcApp := grpcapp.New(log, grpcPort)
	return &App{
		GRPCServer: grpcApp,
	}
}
