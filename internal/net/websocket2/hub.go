package websocket2

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

/*
how does this work?

- hub sits behind an API route: /websocket
- a new connection is established by ServeHTTP
- read and write listen on the connection
- read, reads from connection and broadcasts the messages to hub
- write, writes messages from hub to the connection
- hub's listen method listens on 3 channels. register, unregister and broadcast.
- messages pushed to broadcast are written to all connections by the "write" listener
*/

func read(conn *ConnHandler, hub *Hub) {
	defer func() {
		hub.unregister <- conn
		err := conn.rwc.Close()
		if err != nil {
			slog.Error("conn close: %v", err)
		}
		slog.Debug("read: conn closed")
	}()

	if err := conn.setReadDeadLine(pongWait); err != nil {
		slog.Info("setReadDeadLine: %v\n", err)
		return
	}

	for {
		wsMsg, err := conn.read()
		if err != nil {
			// TODO: handle error
			slog.Error("read err: %v\n", err.Error())
			break
		}

		// TODO: add a way use custom read validation here unsure how yet
		/*
			var envelope msg.Envelope
			slog.Info("read msg: OpCode: %v\n\n", wsMsg.OpCode)
			if err := json.Unmarshal(wsMsg.Payload, &envelope); err != nil {
				slog.Error("wsMsg unmarshal: %v", err)
			} else {
				slog.Error("read msg:\nType: %d\nID: %s\nUserID: %s\n\n", envelope.Typ, envelope.ID, envelope.UserID)
			}
		*/

		hub.broadcast <- wsMsg
	}
}

func write(conn *ConnHandler) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		slog.Debug("write: conn closed")
		if err := conn.rwc.Close(); err != nil {
			slog.Error("error closing connection: %v", err)
		}
	}()

	for {
		select {
		case msg, ok := <-conn.send:
			_ = conn.setWriteDeadLine(writeWait)
			if !ok {
				slog.Error("<-conn.send not ok")
				_ = conn.write(&wsutil.Message{OpCode: ws.OpClose, Payload: []byte{}})
				return
			}

			if err := conn.write(msg); err != nil {
				slog.Error("msg err: %v\n", err)
				return
			}
		case <-ticker.C:
			_ = conn.setWriteDeadLine(writeWait)
			if err := conn.write(&wsutil.Message{OpCode: ws.OpPing, Payload: nil}); err != nil {
				slog.Error("ticker err: %v\n", err)
				return
			}
		}
	}
}

type Hub struct {
	register, unregister chan *ConnHandler
	broadcast            chan *wsutil.Message
	lock                 *sync.Mutex
	connections          map[*ConnHandler]bool
	upgrader             *ws.HTTPUpgrader

	// Capacity of the send channel.
	// If capacity is 0, the send channel is unbuffered.
	Capacity uint
}

// Len returns the number of connections.
func (hub *Hub) Len() int {
	hub.lock.Lock()
	defer hub.lock.Unlock()
	return len(hub.connections)
}

// TODO -- should be able to close all connections via their own channels
func (hub *Hub) Close() error {
	defer func() {
		slog.Info("hub.Close()")
		// close channels
		close(hub.register)
		close(hub.unregister)
		close(hub.broadcast)
	}()

	hub.broadcast <- &wsutil.Message{OpCode: ws.OpClose, Payload: []byte{}} // broadcast close
	return nil
}

/*
NewHub instantiates a new websocket hub.

NOTE: these may be useful to set: Capacity, ReadBufferSize, ReadTimeout, WriteTimeout
*/
func NewHub(cap uint) *Hub {
	hub := &Hub{
		register:    make(chan *ConnHandler),
		unregister:  make(chan *ConnHandler),
		broadcast:   make(chan *wsutil.Message),
		lock:        &sync.Mutex{},
		connections: make(map[*ConnHandler]bool),
		upgrader:    &ws.HTTPUpgrader{
			// TODO: may be fields here that worth setting
		},
		Capacity: cap,
	}

	go hub.listen()
	return hub
}

func (hub *Hub) listen() {
	for {
		select {
		case conn := <-hub.register:
			hub.lock.Lock()
			hub.connections[conn] = true
			hub.lock.Unlock()
		case conn := <-hub.unregister:
			slog.Debug("unregister channel handler")
			delete(hub.connections, conn)
			close(conn.send)
		case msg := <-hub.broadcast:
			for conn := range hub.connections {
				select {
				case conn.send <- msg:
				default:
					// From Gorilla WS
					// https://github.com/gorilla/websocket/tree/master/examples/chat#hub
					// If the client’s send buffer is full, then the hub assumes that the client is dead or stuck. In this case, the hub unregisters the client and closes the websocket
					slog.Debug("conn.send channel buffer possible full\n")
					slog.Debug("broadcast channel handler: default case:\nopCode: %d\npayload: %+v\n", msg.OpCode, msg.Payload)
					close(conn.send)
					delete(hub.connections, conn)
				}
			}
		}
	}
}

func (hub *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO check capacity
	if hub.Capacity > 0 && hub.Len() >= int(hub.Capacity) {
		http.Error(w, "too many connections", http.StatusServiceUnavailable)
		return
	}

	rwc, _, _, err := hub.upgrader.Upgrade(r, w)
	if err != nil {
		// TODO log that there was an error
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	conn := &ConnHandler{
		rwc:  rwc,
		send: make(chan *wsutil.Message, 256), // TODO what's the optimal size?
	}

	hub.register <- conn

	go read(conn, hub)
	go write(conn)
}
