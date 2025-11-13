package websocket

import (
	"context"
	"log"
	"net/http"

	"github.com/coder/websocket"
	"github.com/pmoieni/project-racer-server/internal/net/msg"
)

type Hub struct {
	broadcast   chan *msg.Envelope
	tasks       chan func() error
	subscribers map[*subscriber]struct{}
}

func NewHub() *Hub {
	h := &Hub{
		broadcast:   make(chan *msg.Envelope),
		tasks:       make(chan func() error),
		subscribers: make(map[*subscriber]struct{}),
	}

	log.Println(len(h.subscribers))

	go h.listen()

	return h
}

func (h *Hub) listen() {
	for {
		select {
		case task := <-h.tasks:
			if err := task(); err != nil {
				log.Println(err)
			}
		case msg := <-h.broadcast:
			for s := range h.subscribers {
				s.send <- msg
			}
		}
	}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		// TODO log that there was an error
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sub := newSubscriber(c)

	defer func() {
		if err := h.deleteSubscriber(sub); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}()

	if err := h.addSubscriber(r.Context(), sub); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (h *Hub) addSubscriber(ctx context.Context, s *subscriber) error {
	h.tasks <- func() error {
		h.subscribers[s] = struct{}{}
		return nil
	}

	go func() {
		for msg := range s.send {
			if err := s.write(ctx, msg); err != nil {
				h.tasks <- func() error { return err }
				return
			}
		}
	}()

	for {
		msg, err := s.read(ctx)
		if err != nil {
			return err
		}

		h.broadcast <- msg
	}
}

func (h *Hub) deleteSubscriber(s *subscriber) error {
	if err := s.conn.Close(websocket.StatusNormalClosure, ""); err != nil {
		return err
	}

	h.tasks <- func() error {
		delete(h.subscribers, s)
		return nil
	}

	return nil
}
