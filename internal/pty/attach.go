package pty

import (
	"encoding/binary"
	"net"
	"sync"
	"time"

	"github.com/creack/pty"
)

const greetTimeout = 2 * time.Second

// accept serves each connection. Several may be attached at once and all receive the same
// output: a new attach never evicts another, because connections are served concurrently
// and a late, stale one must not cut off the daemon that just reconnected.
func (h *host) accept(ln net.Listener) {
	var conns sync.WaitGroup
	defer conns.Wait()
	for {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		if !h.track(c) {
			_ = c.Close()
			continue
		}
		conns.Go(func() { h.greet(c) })
	}
}

// track registers c so the host can close it when it stops; false once it is stopping.
func (h *host) track(c net.Conn) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closing {
		return false
	}
	h.conns[c] = struct{}{}
	return true
}

// release forgets c and detaches it.
func (h *host) release(c net.Conn) {
	h.mu.Lock()
	delete(h.conns, c)
	delete(h.attached, c)
	h.mu.Unlock()
	_ = c.Close()
}

// greet reads the first frame: a ping is answered and closed without disturbing the
// attached daemon; an attach takes over the terminal.
func (h *host) greet(c net.Conn) {
	defer h.release(c)
	_ = c.SetReadDeadline(time.Now().Add(greetTimeout))
	kind, _, err := readFrame(c)
	_ = c.SetReadDeadline(time.Time{})
	switch {
	case err != nil:
	case kind == framePing:
		_ = c.SetWriteDeadline(time.Now().Add(greetTimeout))
		_ = writeFrame(c, framePing, nil)
	case kind == frameAttach:
		h.attach(c)
		h.serveConn(c)
	}
}

// attach repaints the new daemon's terminal from the scrollback, then nudges the agent
// to redraw so a full-screen TUI shows its current screen, not just its history.
func (h *host) attach(c net.Conn) {
	h.mu.Lock()
	h.attached[c] = struct{}{}
	ok := h.send(c, frameOutput, h.history.snapshot())
	for _, record := range h.events {
		ok = ok && h.send(c, frameEvent, record)
	}
	exited := h.exited
	if exited && ok {
		h.sendExit(c)
	}
	rows, cols := h.rows, h.cols
	h.mu.Unlock()

	if !exited && cols > 1 {
		_ = pty.Setsize(h.ptmx, &pty.Winsize{Rows: rows, Cols: cols - 1})
		_ = pty.Setsize(h.ptmx, &pty.Winsize{Rows: rows, Cols: cols})
	}
}

// serveConn types the daemon's input until it disconnects.
func (h *host) serveConn(c net.Conn) {
	for {
		kind, p, err := readFrame(c)
		if err != nil {
			return
		}
		h.handle(kind, p)
	}
}

func (h *host) handle(kind byte, p []byte) {
	switch kind {
	case frameInput:
		_, _ = h.ptmx.Write(p) // fails only once the agent is gone; the exit frame says so
	case frameResize:
		if len(p) == 4 {
			h.resize(binary.BigEndian.Uint16(p), binary.BigEndian.Uint16(p[2:]))
		}
	case frameStop:
		h.stop()
	}
}

func (h *host) resize(rows, cols uint16) {
	h.mu.Lock()
	h.rows, h.cols = rows, cols
	h.mu.Unlock()
	_ = pty.Setsize(h.ptmx, &pty.Winsize{Rows: rows, Cols: cols})
}
