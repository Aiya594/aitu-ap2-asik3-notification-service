package model

type EventEnvelope struct {
	EventTime string      `json:"time"`
	Subject   string      `json:"subject"`
	Event     interface{} `json:"event"`
}
