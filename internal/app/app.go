package app

import (
	"log"

	"github.com/Aiya594/aitu-ap2-asik3-notification-service/internal/broker"
	"github.com/Aiya594/aitu-ap2-asik3-notification-service/internal/config"
	"github.com/Aiya594/aitu-ap2-asik3-notification-service/internal/logger"
	"github.com/Aiya594/aitu-ap2-asik3-notification-service/internal/subscrider"
)

type App struct {
	subscriber *subscrider.Subscriber
	broker     *broker.Client
}

func New() *App {
	cfg := config.LoadCfg()

	client, err := broker.New(cfg.NatsURL)
	if err != nil {
		log.Fatal("cannot connect to NATS:", err)
	}

	logg := logger.New()

	sub := subscrider.New(client.Conn, logg)

	return &App{
		subscriber: sub,
		broker:     client,
	}
}

func (a *App) Run() error {
	return a.subscriber.Start()
}
