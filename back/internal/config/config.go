package config

import (
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Key     []byte
	Level   string
	Address string
	Port    int
	DBPath  string
}

const (
	appEnv = "APP_ENV"
)

func Load() (*Config, error) {
	env := os.Getenv(appEnv)
	if env == "" {
		return nil, fmt.Errorf("app env not found: %s", appEnv)
	}

	var cfg Config
	err := envconfig.Process(env, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to process env vars > %w", err)

	}

	return &cfg, nil
}
