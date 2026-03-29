package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DBPath string `envconfig:"DB_PATH" default:"referee.db"`
	Port   int    `envconfig:"PORT" default:"3000"`
}

func Load() (Config, error) {
	_ = godotenv.Load()

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
