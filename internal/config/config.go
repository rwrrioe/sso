package config

import (
	"errors"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type ResendConfig struct {
	From string `yaml:"from"`
	Name string `yaml:"name"`
}

type RedisConfig struct {
	Address  string `yaml:"address"  env:"REDIS_ADDR"  env-required:"true"`
	Password string `yaml:"password" env:"REDIS_PASS"`
	DB       int    `yaml:"db"       env:"REDIS_DB"    env-default:"0"`
	Protocol int    `yaml:"protocol" env-default:"2"`
}

type AuthConfig struct {
	AccessTokenTTL  time.Duration `yaml:"access_token_ttl"  env-default:"15m"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl" env-default:"720h"`
	ResetTokenTTL   time.Duration `yaml:"reset_token_ttl"   env-default:"1h"`
}

type CodeConfig struct {
	CodeTTL time.Duration `yaml:"code_ttl" env-default:"5m"`
}

type Config struct {
	GRPCPort   int    `yaml:"grpc_port" env:"GRPC_PORT" env-default:"9081"`
	LoggerType string `yaml:"logger_type" env-default:"local"`
	Redis      RedisConfig
	Resend     ResendConfig
	Auth       AuthConfig
	Code       CodeConfig
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")

	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		panic("config file does not exist")
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("config path is empty: " + err.Error())
	}

	return &cfg

}
