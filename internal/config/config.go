package config

import (
	"log"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Env  string `env:"ENV" envDefault:"local"`
	GRPC GrpcConfig
	DBConfig
	ImageStorage
}

type GrpcConfig struct {
	Address string        `env:"GRPC_ADDRESS" envDefault:"localhost"`
	Port    string        `env:"GRPC_PORT"`
	Timeout time.Duration `env:"GRPC_TIMEOUT" envDefault:"10s"`
}

type DBConfig struct {
	StoragePath string `env:"STORAGE_PATH"`
}

type ImageStorage struct {
	ServerImageStorage string `env:"PATH_TO_SAVED_IMAGES"`
	ClientImageStorage string `env:"PATH_TO_SAVED_CLIENT"`
}

func MustLoad() *Config {
	cfg := Config{}
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Error parsing .env file: %v", err)
	}
	return &cfg
}
