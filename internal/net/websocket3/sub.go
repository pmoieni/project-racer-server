package websocket3

import (
	"context"

	"github.com/coder/websocket"
)

type subscriber struct {
	conn *websocket.Conn
}

// write has two tasks
// read from connection and write to `ch` channel
// listen on `ch` channel and write to connection only if the ID of `c` of type `*Conn` doesn't match the id of sender
func (s *subscriber) write(ch chan []byte, errc chan error) {
	for msg := range ch {
		if err := s.conn.Write(context.Background(), websocket.MessageBinary, msg); err != nil {
			errc <- err
			return
		}
	}
}

// TODO: don't return on error, use error channel instead
func (s *subscriber) read(ch chan []byte, unregister chan *subscriber, errc chan error) {
	for {
		_, bs, err := s.conn.Read(context.Background())
		if err != nil {
			errc <- err
			return
		}

		ch <- bs
	}
}
