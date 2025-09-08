package engine

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID      string                 `json:"id"`
	Payload interface{}            `json:"payload"`
	Meta    map[string]interface{} `json:"meta"`
	TS      time.Time              `json:"ts"`
}

func NewMessage(payload interface{}, meta map[string]interface{}) *Message {
	if meta == nil {
		meta = map[string]interface{}{}
	}
	return &Message{
		ID:      uuid.NewString(),
		Payload: payload,
		Meta:    meta,
		TS:      time.Now(),
	}
}
