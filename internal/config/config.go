package config

import "os"

type Config struct {
	NatsURL string
}

func LoadCfg() *Config {
	return &Config{NatsURL: os.Getenv("NATS_URL")}
}
