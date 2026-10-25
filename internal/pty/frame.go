package pty

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// Frames on a session socket: one kind byte, a big-endian uint32 length, then the payload.
const (
	frameOutput byte = 'o' // host to daemon: bytes the agent wrote
	frameExit   byte = 'x' // host to daemon: the agent exited; payload is an int32 exit code
	frameInput  byte = 'i' // daemon to host: bytes typed into the terminal
	frameResize byte = 'r' // daemon to host: rows then cols, two uint16
	frameStop   byte = 's' // daemon to host: stop the agent
	frameAttach byte = 'a' // first frame: become the attached daemon
	framePing   byte = 'p' // first frame: only check the host is alive; it answers 'p'
)

// maxFrame bounds a payload so a corrupt length cannot make a peer allocate gigabytes.
const maxFrame = 1 << 20

var errFrameTooLarge = errors.New("frame too large")

func writeFrame(w io.Writer, kind byte, payload []byte) error {
	if len(payload) > maxFrame {
		return errFrameTooLarge
	}
	msg := make([]byte, 5, 5+len(payload))
	msg[0] = kind
	binary.BigEndian.PutUint32(msg[1:], uint32(len(payload))) //nolint:gosec // bounded by maxFrame above
	if _, err := w.Write(append(msg, payload...)); err != nil {
		return fmt.Errorf("write frame %c: %w", kind, err)
	}
	return nil
}

func readFrame(r io.Reader) (byte, []byte, error) {
	var head [5]byte
	if _, err := io.ReadFull(r, head[:]); err != nil {
		return 0, nil, fmt.Errorf("read frame header: %w", err)
	}
	n := binary.BigEndian.Uint32(head[1:])
	if n > maxFrame {
		return 0, nil, errFrameTooLarge
	}
	payload := make([]byte, n)
	if _, err := io.ReadFull(r, payload); err != nil {
		return 0, nil, fmt.Errorf("read frame body: %w", err)
	}
	return head[0], payload, nil
}

// writeChunked sends data larger than one frame as several frames. Empty data still
// sends one empty frame, since some frames (stop) carry no payload.
func writeChunked(w io.Writer, kind byte, data []byte) error {
	for first := true; first || len(data) > 0; first = false {
		n := min(len(data), maxFrame)
		if err := writeFrame(w, kind, data[:n]); err != nil {
			return err
		}
		data = data[n:]
	}
	return nil
}
