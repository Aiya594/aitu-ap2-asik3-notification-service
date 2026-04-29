package app

import (
	"log"
	"os"

	"github.com/Aiya594/aitu-ap2-asik3-notification-service/internal/broker"
	"github.com/Aiya594/aitu-ap2-asik3-notification-service/internal/logger"
	"github.com/Aiya594/aitu-ap2-asik3-notification-service/internal/subscrider"
)

type App struct {
	subscriber *subscrider.Subscriber
	broker     *broker.Client
}

func New() *App {
	url := os.Getenv("NATS_URL")

	client, err := broker.New(url)
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
