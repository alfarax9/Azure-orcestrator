package store

import (
	"sync"

	"github.com/alfarax9/astra/internal/domain"
)

// listeners are the functions told about every change. The map key is only an identity.
type listeners struct {
	mu   sync.Mutex
	next int
	fns  map[int]func(domain.Change)
}

// Subscribe calls fn after every successful write with what changed, until cancel is called.
// fn runs on the goroutine that wrote, so it must not block: hand the change to a channel
// with a non-blocking send.
func (s *Store) Subscribe(fn func(domain.Change)) (cancel func()) {
	s.subs.mu.Lock()
	defer s.subs.mu.Unlock()
	if s.subs.fns == nil {
		s.subs.fns = map[int]func(domain.Change){}
	}
	id := s.subs.next
	s.subs.next++
	s.subs.fns[id] = fn
	return func() {
		s.subs.mu.Lock()
		delete(s.subs.fns, id)
		s.subs.mu.Unlock()
	}
}

// publish tells every subscriber about c.
func (s *Store) publish(c domain.Change) {
	s.subs.mu.Lock()
	fns := make([]func(domain.Change), 0, len(s.subs.fns))
	for _, fn := range s.subs.fns {
		fns = append(fns, fn)
	}
	s.subs.mu.Unlock()
	for _, fn := range fns {
		fn(c)
	}
}
