package websocket

import (
	"context"

	"github.com/coder/websocket"
)

type subscriber struct {
	conn *websocket.Conn
	send chan []byte
}

func newSubscriber(conn *websocket.Conn) *subscriber {
	return &subscriber{conn: conn, send: make(chan []byte)}
}

func (s *subscriber) write(ctx context.Context, bs []byte) error {
	// TODO: don't use JSON
	if err := s.conn.Write(ctx, websocket.MessageBinary, bs); err != nil {
		return err
	}

	return nil
}

func (s *subscriber) read(ctx context.Context) ([]byte, error) {
	_, bs, err := s.conn.Read(ctx)
	if err != nil {
		return nil, err
	}

	return bs, nil
}
