package logger

import (
	"encoding/json"
	"log"
)

type Logger interface {
	Log(subject string, event map[string]interface{}, time string)
}

type JSONLogger struct{}

func New() *JSONLogger {
	return &JSONLogger{}
}

func (l *JSONLogger) Log(subject string, event map[string]interface{}, time string) {
	output := map[string]interface{}{
		"time":    time,
		"subject": subject,
		"event":   event,
	}

	b, err := json.Marshal(output)
	if err != nil {
		log.Printf("failed to marshal log: %v", err)
		return
	}

	log.Println(string(b))
}
