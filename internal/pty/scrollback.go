package pty

// scrollback keeps the last len(buf) bytes the agent wrote, so a daemon that reconnects
// can repaint the terminal. It is a fixed ring: a busy agent never grows the host's memory.
type scrollback struct {
	buf  []byte
	next int  // where the next byte goes
	full bool // buf has wrapped at least once
}

func newScrollback(size int) *scrollback {
	return &scrollback{buf: make([]byte, size)}
}

func (s *scrollback) write(p []byte) {
	size := len(s.buf)
	if len(p) >= size {
		copy(s.buf, p[len(p)-size:])
		s.next, s.full = 0, true
		return
	}
	end := s.next + len(p)
	n := copy(s.buf[s.next:], p)
	copy(s.buf, p[n:])
	if end >= size {
		s.full = true
	}
	s.next = end % size
}

// snapshot returns the kept bytes oldest first, as a copy the caller may keep.
func (s *scrollback) snapshot() []byte {
	if !s.full {
		return append([]byte(nil), s.buf[:s.next]...)
	}
	out := make([]byte, 0, len(s.buf))
	out = append(out, s.buf[s.next:]...)
	return append(out, s.buf[:s.next]...)
}
