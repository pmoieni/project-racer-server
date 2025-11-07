package websocket

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 10 * time.Second
)

type Conn struct {
	id uuid.UUID
	mx *sync.Mutex
	nc net.Conn
	// find the optimal buffer size
	// send chan *wsutil.Message
}

func newConn(nc net.Conn) (*Conn, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	return &Conn{
		id: id,
		mx: &sync.Mutex{},
		nc: nc,
	}, nil
}

func (c *Conn) setWriteDeadLine(d time.Duration) error {
	return c.nc.SetWriteDeadline(time.Now().Add(d))
}

func (c *Conn) setReadDeadLine(d time.Duration) error {
	return c.nc.SetReadDeadline(time.Now().Add(d))
}

func (c *Conn) controlHandler(h ws.Header) error {
	switch op := h.OpCode; op {
	case ws.OpPing:
		return c.handlePing(h)
	case ws.OpPong:
		return c.handlePong(h)
	case ws.OpClose:
		return c.handleClose(h)
	}

	return wsutil.ErrNotControlFrame
}

func (c *Conn) handlePing(h ws.Header) error {
	slog.Info("ping")
	return nil
}

func (c *Conn) handlePong(h ws.Header) error {
	slog.Info("pong")
	return c.setReadDeadLine(pongWait)
}

func (c *Conn) handleClose(h ws.Header) error {
	slog.Info("close")
	return nil
}

// write has two tasks
// read from connection and write to `ch` channel
// listen on `ch` channel and write to connection only if the ID of `c` of type `*Conn` doesn't match the id of sender
func (c *Conn) write(ch chan *wsutil.Message, errc chan error) {
	for msg := range ch {
		if err := wsutil.WriteServerMessage(c.nc, msg.OpCode, msg.Payload); err != nil {
			errc <- err
			return
		}
		// connection-specific message handling
		/*
			case msg := <-c.send:
				if err := wsutil.WriteServerMessage(c.nc, msg.OpCode, msg.Payload); err != nil {
					errc <- err
					return
				}
		*/
	}
}

// TODO: don't return on error, use error channel instead
func (c *Conn) read(ch chan *wsutil.Message, unregister chan uuid.UUID, errc chan error) {
	defer func() {
		unregister <- c.id
		err := c.nc.Close()
		if err != nil {
			errc <- err
			// slog.Error("conn close: %v", err)
		}
		slog.Debug("read: conn closed")
	}()

	if err := c.setReadDeadLine(pongWait); err != nil {
		errc <- err
		return
	}

	cr := wsutil.NewServerSideReader(c.nc)

	for {
		header, err := cr.NextFrame()
		if err != nil {
			errc <- c.formatErr(err)
			return
		}

		if header.OpCode.IsControl() {
			if err := c.controlHandler(header); err != nil {
				errc <- c.formatErr(err)
				return
			}
			continue
		}

		// where want = ws.OpText|ws.OpBinary
		// NOTE -- eq: h.OpCode != 0 && h.OpCode != want
		if want := (ws.OpText | ws.OpBinary); header.OpCode&want == 0 {
			if err := cr.Discard(); err != nil {
				errc <- c.formatErr(err)
				return
			}
			continue
		}

		p, err := io.ReadAll(cr)
		if err != nil {
			errc <- c.formatErr(err)
			return
		}

		ch <- &wsutil.Message{OpCode: header.OpCode, Payload: p}
	}

}

func (c *Conn) formatErr(err error) error {
	return fmt.Errorf("conn [%s]: %v", c.id.String(), err)
}
