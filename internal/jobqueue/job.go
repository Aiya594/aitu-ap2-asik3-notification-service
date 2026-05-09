package jobqueue

// Job represents a background notification task.
type Job struct {
	IdempotencyKey string `json:"idempotency_key"`
	AppointmentID  string `json:"appointment_id"`
	DoctorID       string `json:"doctor_id"`
	OccurredAt     string `json:"occurred_at"`
	Channel        string `json:"channel"`
	Recipient      string `json:"recipient"`
	Message        string `json:"message"`
}
