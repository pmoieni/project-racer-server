package msg

import (
	"encoding/json"

	"github.com/google/uuid"
)

type (
	MsgType   int
	NoteState int

	Envelope struct {
		// Message identifier
		ID  uuid.UUID `json:"id"`
		Typ MsgType   `json:"type"`
		// client identifier
		UserID uuid.UUID `json:"userId"`
		// Actual message data.
		Payload json.RawMessage `json:"payload"`
	}

	TextMsg struct {
		DisplayName string `json:"displayName"`
		Body        string `json:"body"`
	}

	ConnectMsg struct {
		UserID   uuid.UUID `json:"userId"`
		UserName string    `json:"userName"`
	}
)

const (
	TEXT MsgType = iota
	CONNECT
)

func (e *Envelope) SetPayload(payload any) error {
	p, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	e.Payload = p
	return nil
}

func (e *Envelope) Unwrap(msg any) error {
	return json.Unmarshal(e.Payload, msg)
}
