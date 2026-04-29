package subscrider

type EventEnvelope struct {
	EventType  string                 `json:"event_type"`
	OccurredAt string                 `json:"occurred_at"`
	Payload    map[string]interface{} `json:"-"`
}
