package subscrider

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Aiya594/aitu-ap2-asik3-notification-service/internal/logger"
	"github.com/Aiya594/aitu-ap2-asik3-notification-service/internal/model"
	"github.com/nats-io/nats.go"
)

type Subscriber struct {
	conn   *nats.Conn
	logger logger.Logger
}

func New(conn *nats.Conn, logger logger.Logger) *Subscriber {
	return &Subscriber{
		conn:   conn,
		logger: logger,
	}
}

func (s *Subscriber) Start() error {
	subjects := []string{
		"doctors.created",
		"appointments.created",
		"appointments.status_updated",
	}

	for _, subject := range subjects {
		_, err := s.conn.Subscribe(subject, s.handleMessage)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Subscriber) handleMessage(msg *nats.Msg) {

	var payload map[string]interface{}

	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		log.Printf("invalid message subject=%s err=%v", msg.Subject, err)
		return
	}

	event := model.EventEnvelope{
		EventTime: time.Now().UTC().Format(time.RFC3339),
		Subject:   msg.Subject,
		Event:     payload,
	}

	s.logger.Log(event)
}
