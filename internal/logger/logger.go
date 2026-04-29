package logger

import (
	"encoding/json"
	"log"

	"github.com/Aiya594/aitu-ap2-asik3-notification-service/internal/model"
)

type Logger interface {
	Log(output model.EventEnvelope) error
}

type JSONLogger struct{}

func New() *JSONLogger {
	return &JSONLogger{}
}

func (l *JSONLogger) Log(output model.EventEnvelope) error {
	b, err := json.Marshal(output)
	if err != nil {
		log.Printf("failed to marshal log: %v", err)
		return err
	}

	log.Println(string(b))
	return nil
}
