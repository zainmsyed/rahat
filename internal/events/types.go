package events

import "time"

type EventLog struct {
	ID           string
	UserID       string
	OccurrenceID string
	Channel      string
	EventType    string
	MessageType  string
	PayloadJSON  string
	OccurredAt   time.Time
}
