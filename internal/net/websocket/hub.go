package websocket

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
)

const (
	pingPeriod = (pongWait * 9) / 10
)

type Hub struct {
	mx         *sync.Mutex
	register   chan *Conn
	unregister chan uuid.UUID
	// should only allow messages with `ws.OpText` or `ws.OpBinary`
	broadcast   chan *wsutil.Message
	errc        chan error
	connections map[uuid.UUID]*Conn
	upgrader    *ws.HTTPUpgrader

	// Capacity of the send channel.
	// If capacity is 0, the send channel is unbuffered.
	Capacity uint
}

func NewHub(cap uint) *Hub {
	hub := &Hub{
		mx:          &sync.Mutex{},
		register:    make(chan *Conn),
		unregister:  make(chan uuid.UUID),
		broadcast:   make(chan *wsutil.Message),
		errc:        make(chan error),
		connections: make(map[uuid.UUID]*Conn),
		upgrader:    &ws.HTTPUpgrader{},
		Capacity:    cap,
	}

	go hub.listen()

	return hub
}

// Len returns the number of connections.
func (h *Hub) Len() int {
	h.mx.Lock()
	defer h.mx.Unlock()
	return len(h.connections)
}

// listen handles client register, unregister, pinging and errors
func (h *Hub) listen() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case conn := <-h.register:
			h.mx.Lock()
			h.connections[conn.id] = conn
			h.mx.Unlock()
		case cid := <-h.unregister:
			h.mx.Lock()
			// close(conn.send)
			delete(h.connections, cid)
			h.mx.Unlock()
		case err := <-h.errc:
			log.Println(err)
		case <-ticker.C:
			h.broadcast <- &wsutil.Message{OpCode: ws.OpPing, Payload: nil}
		}
	}
}

/*
func (h *Hub) broadcastMsg(msg *wsutil.Message) {
	for conn := range h.connections {
		select {
		case conn.send <- msg:
		default:
			// From Gorilla WS
			// https://github.com/gorilla/websocket/tree/master/examples/chat#hub
			// If the client’s send buffer is full, then the hub assumes that the client is dead or stuck. In this case, the hub unregisters the client and closes the websocket
			slog.Debug("conn.send channel buffer possible full\n")
			slog.Debug("broadcast channel handler: default case:\nopCode: %d\npayload: %+v\n", msg.OpCode, msg.Payload)
			h.mx.Lock()
			close(conn.send)
			delete(h.connections, conn)
			h.mx.Unlock()
		}
	}
}
*/

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO check capacity
	if h.Capacity > 0 && h.Len() >= int(h.Capacity) {
		http.Error(w, "too many connections", http.StatusServiceUnavailable)
		return
	}

	nc, _, _, err := h.upgrader.Upgrade(r, w)
	if err != nil {
		// TODO log that there was an error
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	conn, err := newConn(nc)
	if err != nil {
		http.Error(w, "failed to establish connection", http.StatusInternalServerError)
		return
	}

	go conn.read(h.broadcast, h.unregister, h.errc)
	go conn.write(h.broadcast, h.errc)

	h.register <- conn

}
