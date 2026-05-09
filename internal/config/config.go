package config

import "os"

type Config struct {
	NatsURL        string
	RedisURL       string
	GatewayURL     string
	WorkerPoolSize string
}

func LoadCfg() *Config {
	return &Config{
		NatsURL:        os.Getenv("NATS_URL"),
		RedisURL:       os.Getenv("REDIS_URL"),
		GatewayURL:     os.Getenv("GATEWAY_URL"),
		WorkerPoolSize: os.Getenv("WORKER_POOL_SIZE"),
	}
}
