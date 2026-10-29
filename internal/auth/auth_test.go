package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/alfarax9/astra/internal/domain"
)

// memory is a store in maps that follows the real one's rules.
type memory struct {
	pairings map[string]time.Time
	devices  map[string]domain.Device // by token hash
	touched  int
}

func newMemory() *memory {
	return &memory{pairings: map[string]time.Time{}, devices: map[string]domain.Device{}}
}

func (m *memory) CreatePairing(_ context.Context, h string, exp time.Time) error {
	m.pairings[h] = exp
	return nil
}

func (m *memory) ConsumePairing(_ context.Context, h string, now time.Time) error {
	exp, ok := m.pairings[h]
	delete(m.pairings, h)
	if !ok || !exp.After(now) {
		return domain.ErrNotFound
	}
	return nil
}

func (m *memory) CreateDevice(_ context.Context, d domain.Device, h string) error {
	m.devices[h] = d
	return nil
}

func (m *memory) DeviceByToken(_ context.Context, h string) (domain.Device, error) {
	if d, ok := m.devices[h]; ok {
		return d, nil
	}
	return domain.Device{}, domain.ErrNotFound
}

func (m *memory) TouchDevice(_ context.Context, id string, at time.Time) error {
	for h, d := range m.devices {
		if d.ID == id {
			d.LastSeen = domain.Timestamp{Time: at}
			m.devices[h] = d
			m.touched++
		}
	}
	return nil
}

func fixture() (*Service, *memory, *time.Time) {
	m := newMemory()
	clock := time.UnixMilli(1700000000000)
	return New(m, m, func() time.Time { return clock }), m, &clock
}

func TestPairingMakesADeviceTokenOnce(t *testing.T) {
	ctx := t.Context()
	s, m, _ := fixture()
	pairing, err := s.NewPairing(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(pairing.String(), pairing.Reveal()) {
		t.Error("the pairing token prints its value")
	}
	for h := range m.pairings {
		if h == pairing.Reveal() {
			t.Error("the pairing token was stored in the clear")
		}
	}
	dev, token, err := s.Pair(ctx, pairing.Reveal(), "  Laptop  ")
	if err != nil || dev.Name != "Laptop" || dev.ID == "" {
		t.Fatalf("Pair = %+v, %v", dev, err)
	}
	for h := range m.devices {
		if h == token.Reveal() {
			t.Error("the device token was stored in the clear")
		}
	}
	if _, _, err := s.Pair(ctx, pairing.Reveal(), "Laptop"); !errors.Is(err, ErrUnauthorized) {
		t.Errorf("a pairing token used twice: err = %v", err)
	}
	got, err := s.Authenticate(ctx, token.Reveal())
	if err != nil || got.ID != dev.ID {
		t.Errorf("Authenticate = %+v, %v", got, err)
	}
}

func TestPairingTokensExpire(t *testing.T) {
	ctx := t.Context()
	s, _, clock := fixture()
	pairing, _ := s.NewPairing(ctx)
	*clock = clock.Add(PairingTTL)
	if _, _, err := s.Pair(ctx, pairing.Reveal(), "Late"); !errors.Is(err, ErrUnauthorized) {
		t.Errorf("an expired pairing token: err = %v", err)
	}
}

func TestBadPairingAttemptsLearnNothing(t *testing.T) {
	ctx := t.Context()
	s, _, _ := fixture()
	if _, _, err := s.Pair(ctx, "guess", "Phone"); !errors.Is(err, ErrUnauthorized) {
		t.Errorf("a wrong token: err = %v", err)
	}
	if _, _, err := s.Pair(ctx, "", "Phone"); !errors.Is(err, ErrUnauthorized) {
		t.Errorf("an empty token: err = %v", err)
	}
	if _, _, err := s.Pair(ctx, strings.Repeat("x", maxToken+1), "Phone"); !errors.Is(err, ErrUnauthorized) {
		t.Errorf("a huge token: err = %v", err)
	}
	pairing, _ := s.NewPairing(ctx)
	for _, name := range []string{"", "  ", strings.Repeat("n", maxName+1), "bad\x00name", "line\nbreak"} {
		if _, _, err := s.Pair(ctx, pairing.Reveal(), name); !errors.Is(err, ErrInvalid) {
			t.Errorf("name %q: err = %v, want ErrInvalid", name, err)
		}
	}
	// Refused names do not spend the token: only a token that was tried is spent.
	if _, _, err := s.Pair(ctx, pairing.Reveal(), "Phone"); err != nil {
		t.Errorf("the pairing token should still work after refused names: %v", err)
	}
}

func TestAuthenticateRefusesUnknownTokensAndTracksLastSeen(t *testing.T) {
	ctx := t.Context()
	s, m, clock := fixture()
	if _, err := s.Authenticate(ctx, "nope"); !errors.Is(err, ErrUnauthorized) {
		t.Errorf("unknown token: err = %v", err)
	}
	if _, err := s.Authenticate(ctx, ""); !errors.Is(err, ErrUnauthorized) {
		t.Errorf("empty token: err = %v", err)
	}
	pairing, _ := s.NewPairing(ctx)
	_, token, _ := s.Pair(ctx, pairing.Reveal(), "Laptop")
	if _, err := s.Authenticate(ctx, token.Reveal()); err != nil || m.touched != 0 {
		t.Fatalf("a request right after pairing: %v, touched %d", err, m.touched)
	}
	*clock = clock.Add(2 * touchEvery)
	d, err := s.Authenticate(ctx, token.Reveal())
	if err != nil || m.touched != 1 || d.LastSeen.UnixMilli() != clock.UnixMilli() {
		t.Errorf("after a while: %+v, %v, touched %d", d, err, m.touched)
	}
	if _, err := s.Authenticate(ctx, token.Reveal()); err != nil || m.touched != 1 {
		t.Error("a busy device must not be written on every request")
	}
}
