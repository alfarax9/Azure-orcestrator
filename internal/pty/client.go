package pty

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"
)

// Conn is the daemon's end of a session socket.
type Conn struct {
	conn net.Conn
	mu   sync.Mutex // serialises writes from several callers
}

// Event is one thing the session reports: output the agent wrote, or its exit.
type Event struct {
	Output   []byte
	Exited   bool
	ExitCode int
}

// Dial attaches to the session host listening on socket, replacing any daemon attached
// before. The first events repaint the terminal from the host's scrollback.
func Dial(ctx context.Context, socket string) (*Conn, error) {
	var d net.Dialer
	c, err := d.DialContext(ctx, "unix", socket)
	if err != nil {
		return nil, fmt.Errorf("attach %s: %w", socket, err)
	}
	if err := writeFrame(c, frameAttach, nil); err != nil {
		_ = c.Close()
		return nil, err
	}
	return &Conn{conn: c}, nil
}

// ping checks that a host answers on socket without attaching to it.
func ping(ctx context.Context, socket string) error {
	d := net.Dialer{Timeout: time.Second}
	c, err := d.DialContext(ctx, "unix", socket)
	if err != nil {
		return fmt.Errorf("ping %s: %w", socket, err)
	}
	defer func() { _ = c.Close() }()
	_ = c.SetDeadline(time.Now().Add(greetTimeout))
	if err := writeFrame(c, framePing, nil); err != nil {
		return err
	}
	if kind, _, err := readFrame(c); err != nil || kind != framePing {
		return fmt.Errorf("ping %s: no answer", socket)
	}
	return nil
}

// Next blocks until the session reports output or its exit.
func (c *Conn) Next() (Event, error) {
	for {
		kind, p, err := readFrame(c.conn)
		if err != nil {
			return Event{}, err
		}
		switch kind {
		case frameOutput:
			return Event{Output: p}, nil
		case frameExit:
			if len(p) == 4 {
				return Event{Exited: true, ExitCode: int(int32(binary.BigEndian.Uint32(p)))}, nil //nolint:gosec // int32 sent by the host
			}
		}
	}
}

// Write sends keyboard input to the agent.
func (c *Conn) Write(p []byte) (int, error) {
	if err := c.send(frameInput, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Resize tells the agent the terminal's new size.
func (c *Conn) Resize(rows, cols uint16) error {
	var p [4]byte
	binary.BigEndian.PutUint16(p[:], rows)
	binary.BigEndian.PutUint16(p[2:], cols)
	return c.send(frameResize, p[:])
}

// Stop asks the host to hang up the agent; Next then reports its exit.
func (c *Conn) Stop() error {
	return c.send(frameStop, nil)
}

// SetReadDeadline bounds how long Next waits.
func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

// Close detaches from the session. The agent keeps running.
func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) send(kind byte, p []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return writeChunked(c.conn, kind, p)
}
