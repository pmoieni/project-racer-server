package websocket2

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

// using websocket for now, will be switching to Quic and WebTransport later
type ConnHandler struct {
	rwc net.Conn

	send chan *wsutil.Message
}

func (c *ConnHandler) setWriteDeadLine(d time.Duration) error {
	return c.rwc.SetWriteDeadline(time.Now().Add(d))
}

func (c *ConnHandler) setReadDeadLine(d time.Duration) error {
	return c.rwc.SetReadDeadline(time.Now().Add(d))
}

func (c *ConnHandler) read() (*wsutil.Message, error) {
	r := wsutil.NewServerSideReader(c.rwc)

	for {
		h, err := r.NextFrame()
		if err != nil {
			return nil, fmt.Errorf("next frame: %w", err)
		}

		if h.OpCode.IsControl() {
			if err := c.controlHandler(h); err != nil {
				return nil, fmt.Errorf("control handler: %w", err)
			}
			continue
		}

		/*
			// TODO check if this worth doing
			if !h.OpCode.IsData() {
				if h.OpCode.IsControl() {
					if err := c.controlHandler(h, r); err != nil {
						return nil, fmt.Errorf("control handler: %w", err)
					}
					continue
				}
			 	if err := r.Discard(); err != nil {
			 		return nil, fmt.Errorf("discard: %w", err)
			 	}
			 	continue
			}
		*/

		// where want = ws.OpText|ws.OpBinary
		// NOTE -- eq: h.OpCode != 0 && h.OpCode != want
		if want := (ws.OpText | ws.OpBinary); h.OpCode&want == 0 {
			if err := r.Discard(); err != nil {
				return nil, fmt.Errorf("discard: %w", err)
			}
			continue
		}

		// TODO the custom handler to parse payload could be done here (?)

		p, err := io.ReadAll(r)
		if err != nil {
			return nil, fmt.Errorf("read all: %w", err)
		}
		return &wsutil.Message{OpCode: h.OpCode, Payload: p}, nil
	}
}

func (c *ConnHandler) write(msg *wsutil.Message) error {
	frame := ws.NewFrame(msg.OpCode, true, msg.Payload)
	return ws.WriteFrame(c.rwc, frame)
}

func (c *ConnHandler) controlHandler(h ws.Header) error {
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

func (c *ConnHandler) handlePing(h ws.Header) error {
	slog.Info("ping")
	return nil
}

func (c *ConnHandler) handlePong(h ws.Header) error {
	slog.Info("pong")
	return c.setReadDeadLine(pongWait)
}

func (c *ConnHandler) handleClose(h ws.Header) error {
	slog.Info("close")
	return nil
}
