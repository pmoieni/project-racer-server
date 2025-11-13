package websocket

import (
	"context"

	"github.com/coder/websocket"
	"github.com/pmoieni/project-racer-server/internal/net/msg"
)

type subscriber struct {
	conn *websocket.Conn
	send chan *msg.Envelope
}

func newSubscriber(conn *websocket.Conn) *subscriber {
	return &subscriber{conn: conn, send: make(chan *msg.Envelope)}
}

func (s *subscriber) write(ctx context.Context, msg *msg.Envelope) error {
	// TODO: don't use JSON
	if err := s.conn.Write(ctx, websocket.MessageBinary, msg.Payload); err != nil {
		return err
	}

	return nil
}

func (s *subscriber) read(ctx context.Context) (*msg.Envelope, error) {
	_, bs, err := s.conn.Read(ctx)
	if err != nil {
		return nil, err
	}

	return &msg.Envelope{Typ: msg.TEXT, Payload: bs}, nil
}
