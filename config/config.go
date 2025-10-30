package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string     `yaml:"env"         env-required:"true"`
	LogLevel   string     `yaml:"log_level"                       env-default:"info"`
	HttpServer httpServer `yaml:"http_server"`
}

type httpServer struct {
	Address     string        `yaml:"address"      env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout"      env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		slog.Error("CONFIG_PATH is not set")
		os.Exit(1)
	}

	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		slog.Error("config file does not exist", "config_path", configPath)
		os.Exit(1)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		slog.Error("cannot read config", "error", err)
	}

	return &cfg
}
