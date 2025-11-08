package websocket

import (
	"context"
	"encoding/json"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/pmoieni/project-racer-server/internal/net/msg"
)

type subscriber struct {
	id   uuid.UUID
	conn *websocket.Conn
	send chan *msg.Envelope
}

func (s *subscriber) write(ctx context.Context, tasks chan func() error) {
	// TODO: don't use JSON
	for msg := range s.send {
		bs, err := json.Marshal(msg)
		if err != nil {
			tasks <- func() error {
				return err
			}
		}

		if err := s.conn.Write(ctx, websocket.MessageBinary, bs); err != nil {
			tasks <- func() error {
				return s.conn.Close(websocket.StatusGoingAway, "")
			}
		}
	}
}

func (s *subscriber) read(ctx context.Context) (*msg.Envelope, error) {
	_, bs, err := s.conn.Read(ctx)
	if err != nil {
		return nil, err
	}

	var msg *msg.Envelope
	if err := json.Unmarshal(bs, msg); err != nil {
		return nil, err
	}

	return msg, nil
}
