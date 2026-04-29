package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Aiya594/aitu-ap2-asik3-notification-service/internal/app"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	app := app.New()

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}

	log.Println("Notification service started")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	<-sig
	log.Println("shutting down")
}
