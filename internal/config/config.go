package config

import "os"

type Config struct {
	Port    string
	NatsURL string
}

func LoadCfg() *Config {
	return &Config{Port: os.Getenv("PORT"), NatsURL: os.Getenv("NATS_URL")}
}
