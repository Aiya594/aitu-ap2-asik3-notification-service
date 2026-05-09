package subscrider

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Aiya594/aitu-ap2-asik3-notification-service/internal/jobqueue"
	"github.com/Aiya594/aitu-ap2-asik3-notification-service/internal/logger"
	"github.com/Aiya594/aitu-ap2-asik3-notification-service/internal/model"
	"github.com/nats-io/nats.go"
)

type Subscriber struct {
	conn   *nats.Conn
	logger logger.Logger
	pool   *jobqueue.WorkerPool
}

func New(conn *nats.Conn, l logger.Logger, pool *jobqueue.WorkerPool) *Subscriber {
	return &Subscriber{conn: conn, logger: l, pool: pool}
}

func (s *Subscriber) Start() error {
	subjects := []string{
		"doctors.created",
		"appointments.created",
		"appointments.status_updated",
	}

	for _, subject := range subjects {
		subj := subject // capture
		_, err := s.conn.Subscribe(subj, func(msg *nats.Msg) {
			s.handleMessage(msg)
		})
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

	// Enqueue job only for appointments.status_updated with new_status = "done"
	if msg.Subject == "appointments.status_updated" {
		newStatus, _ := payload["New_status"].(string)
		if newStatus == "" {
			newStatus, _ = payload["new_status"].(string)
		}
		if newStatus == "done" {
			s.enqueueNotificationJob(payload)
		}
	}
}

func (s *Subscriber) enqueueNotificationJob(payload map[string]interface{}) {
	appointmentID, _ := payload["ID"].(string)
	if appointmentID == "" {
		appointmentID, _ = payload["id"].(string)
	}
	doctorID, _ := payload["Doctor_id"].(string)
	if doctorID == "" {
		doctorID, _ = payload["doctor_id"].(string)
	}
	occurredAt, _ := payload["Occurred_at"].(string)
	if occurredAt == "" {
		// try to get as RFC3339 from interface
		if t, ok := payload["occurred_at"].(string); ok {
			occurredAt = t
		} else {
			occurredAt = time.Now().UTC().Format(time.RFC3339)
		}
	}

	eventType := "appointments.status_updated"
	raw := fmt.Sprintf("%s%s%s", eventType, appointmentID, occurredAt)
	hash := sha256.Sum256([]byte(raw))
	idempotencyKey := fmt.Sprintf("%x", hash)

	job := jobqueue.Job{
		IdempotencyKey: idempotencyKey,
		AppointmentID:  appointmentID,
		DoctorID:       doctorID,
		OccurredAt:     occurredAt,
		Channel:        "email",
		Recipient:      "patient@clinic.kz",
		Message:        fmt.Sprintf("Your appointment %s with doctor %s is complete.", appointmentID, doctorID),
	}

	s.pool.Enqueue(context.Background(), job)
}
