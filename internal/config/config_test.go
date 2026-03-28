package config

import (
	"os"
	"testing"
	"time"
)

func createTempConfig(t *testing.T, content string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "test_config_*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	if _, err = tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}
	tmpFile.Close()

	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})

	return tmpFile.Name()
}

func TestMustLoad(t *testing.T) {
	path := createTempConfig(t, `
grpc_port: 9081
logger_type: local

redis:
  address: "localhost:6379"
  password: ""
  db: 0
  protocol: 2

resend:
  from: "test@example.com"
  name: "TestApp"

auth:
  access_token_ttl: 15m
  refresh_token_ttl: 720h
  reset_token_ttl: 1h

code:
  code_ttl: 5m
`)

	t.Setenv("CONFIG_PATH", path)
	t.Setenv("REDIS_ADDR", "localhost:6379")

	cfg := MustLoad()

	if cfg.GRPCPort != 9081 {
		t.Errorf("expected GRPCPort 9081, got %d", cfg.GRPCPort)
	}
	if cfg.LoggerType != "local" {
		t.Errorf("expected LoggerType local, got %s", cfg.LoggerType)
	}
	if cfg.Redis.Address != "localhost:6379" {
		t.Errorf("expected Redis.Address localhost:6379, got %s", cfg.Redis.Address)
	}
	if cfg.Redis.Protocol != 2 {
		t.Errorf("expected Redis.Protocol 2, got %d", cfg.Redis.Protocol)
	}
	if cfg.Resend.From != "test@example.com" {
		t.Errorf("expected Resend.From test@example.com, got %s", cfg.Resend.From)
	}
	if cfg.Auth.AccessTokenTTL != 15*time.Minute {
		t.Errorf("expected AccessTokenTTL 15m, got %v", cfg.Auth.AccessTokenTTL)
	}
	if cfg.Auth.RefreshTokenTTL != 720*time.Hour {
		t.Errorf("expected RefreshTokenTTL 720h, got %v", cfg.Auth.RefreshTokenTTL)
	}
	if cfg.Code.CodeTTL != 5*time.Minute {
		t.Errorf("expected CodeTTL 5m, got %v", cfg.Code.CodeTTL)
	}
}

func TestMustLoad_PanicOnMissingFile(t *testing.T) {
	t.Setenv("CONFIG_PATH", "nonexistent.yaml")

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic, got none")
		}
	}()

	MustLoad()
}

func TestMustLoad_EnvOverridesYaml(t *testing.T) {
	path := createTempConfig(t, `
grpc_port: 9081

redis:
address: "localhost:6379"
protocol: 2
`)

	t.Setenv("CONFIG_PATH", path)
	t.Setenv("REDIS_ADDR", "redis:6379")
	t.Setenv("GRPC_PORT", "9090")

	cfg := MustLoad()

	if cfg.Redis.Address != "redis:6379" {
		t.Errorf("expected Redis.Address redis:6379, got %s", cfg.Redis.Address)
	}
	if cfg.GRPCPort != 9090 {
		t.Errorf("expected GRPCPort 9090, got %d", cfg.GRPCPort)
	}
}
