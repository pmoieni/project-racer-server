package websocket3

import (
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"golang.org/x/time/rate"
)

type Hub struct {
	unregister chan *subscriber
	// should only allow messages with `ws.OpText` or `ws.OpBinary`
	broadcast                   chan []byte
	errc                        chan error
	subscriberMessageBufferSize uint
	subscribersMx               *sync.Mutex
	subscribers                 map[*subscriber]struct{}
	publishLimiter              *rate.Limiter
}

func NewHub() *Hub {
	return &Hub{
		unregister:                  make(chan *subscriber),
		broadcast:                   make(chan []byte),
		errc:                        make(chan error),
		subscriberMessageBufferSize: 16,
		subscribersMx:               &sync.Mutex{},
		subscribers:                 make(map[*subscriber]struct{}),
		publishLimiter:              rate.NewLimiter(rate.Every(time.Millisecond*100), 8),
	}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		// TODO log that there was an error
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sub := &subscriber{conn: c}

	h.addSubscriber(sub)
	defer h.deleteSubscriber(sub)

	go sub.read(h.broadcast, h.unregister, h.errc)
	go sub.write(h.broadcast, h.errc)
}

func (h *Hub) addSubscriber(s *subscriber) {
	h.subscribersMx.Lock()
	h.subscribers[s] = struct{}{}
	h.subscribersMx.Unlock()
}

func (h *Hub) deleteSubscriber(s *subscriber) {
	h.subscribersMx.Lock()
	delete(h.subscribers, s)
	h.subscribersMx.Unlock()
}
