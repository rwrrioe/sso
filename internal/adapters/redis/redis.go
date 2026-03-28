package redis

import (
	"log/slog"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Address  string
	Password string
	DB       int
	Protocol int
}

type Storage struct {
	log    *slog.Logger
	client *redis.Client
}

func New(log *slog.Logger, cfg *Config) (*Storage, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
		Protocol: cfg.Protocol,
	})

	return &Storage{
		client: rdb,
		log:    log,
	}, nil
}
