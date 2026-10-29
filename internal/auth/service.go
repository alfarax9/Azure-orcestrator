package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/alfarax9/astra/internal/domain"
)

// Pairings keeps the hashes of pairing tokens that have not been used.
type Pairings interface {
	CreatePairing(ctx context.Context, hash string, expires time.Time) error
	ConsumePairing(ctx context.Context, hash string, now time.Time) error
}

// Devices keeps paired devices by the hash of their token.
type Devices interface {
	CreateDevice(ctx context.Context, d domain.Device, hash string) error
	DeviceByToken(ctx context.Context, hash string) (domain.Device, error)
	TouchDevice(ctx context.Context, id string, at time.Time) error
}

// Errors the API turns into answers.
var (
	// ErrUnauthorized means the token is missing, wrong, expired, used or revoked. It never
	// says which, so a caller learns nothing from guessing.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrInvalid means the device name was refused.
	ErrInvalid = errors.New("invalid")
)

const (
	// PairingTTL is how long a pairing token can be used after it was made.
	PairingTTL = 10 * time.Minute
	// touchEvery is how often a busy device's last-seen time is written.
	touchEvery = time.Minute

	tokenBytes = 32
	maxToken   = 128
	maxName    = 60
)

// Service pairs devices and authenticates them.
type Service struct {
	pairings Pairings
	devices  Devices
	now      func() time.Time
}

// New returns a Service. now is the clock; tests pass a fixed one.
func New(p Pairings, d Devices, now func() time.Time) *Service {
	return &Service{pairings: p, devices: d, now: now}
}

// NewPairing makes a pairing token that works once, for PairingTTL. The caller hands it to
// the client and never shows it.
func (s *Service) NewPairing(ctx context.Context) (domain.Secret, error) {
	token, err := newToken()
	if err != nil {
		return domain.Secret{}, err
	}
	if err := s.pairings.CreatePairing(ctx, hash(token), s.now().Add(PairingTTL)); err != nil {
		return domain.Secret{}, err
	}
	return domain.NewSecret(token), nil
}

// Pair trades a pairing token for a device token. A refused name is reported before the
// token is looked at, so it spends nothing; a token that was tried is used up, right or wrong.
func (s *Service) Pair(ctx context.Context, pairing, name string) (domain.Device, domain.Secret, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > maxName || strings.ContainsFunc(name, unicode.IsControl) {
		return domain.Device{}, domain.Secret{}, fmt.Errorf("%w: a device name is 1 to %d characters without control characters", ErrInvalid, maxName)
	}
	if len(pairing) == 0 || len(pairing) > maxToken {
		return domain.Device{}, domain.Secret{}, ErrUnauthorized
	}
	switch err := s.pairings.ConsumePairing(ctx, hash(pairing), s.now()); {
	case errors.Is(err, domain.ErrNotFound):
		return domain.Device{}, domain.Secret{}, ErrUnauthorized
	case err != nil:
		return domain.Device{}, domain.Secret{}, err
	}
	token, err := newToken()
	if err != nil {
		return domain.Device{}, domain.Secret{}, err
	}
	now := domain.Timestamp{Time: s.now()}
	d := domain.Device{ID: domain.NewDeviceID(), Name: name, CreatedAt: now, LastSeen: now}
	if err := s.devices.CreateDevice(ctx, d, hash(token)); err != nil {
		return domain.Device{}, domain.Secret{}, err
	}
	return d, domain.NewSecret(token), nil
}

// Authenticate finds the device a token belongs to.
func (s *Service) Authenticate(ctx context.Context, token string) (domain.Device, error) {
	if len(token) == 0 || len(token) > maxToken {
		return domain.Device{}, ErrUnauthorized
	}
	d, err := s.devices.DeviceByToken(ctx, hash(token))
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return domain.Device{}, ErrUnauthorized
	case err != nil:
		return domain.Device{}, err
	}
	if now := s.now(); now.Sub(d.LastSeen.Time) >= touchEvery {
		// A failure only leaves the old time: the request is still fine.
		if s.devices.TouchDevice(ctx, d.ID, now) == nil {
			d.LastSeen = domain.Timestamp{Time: now}
		}
	}
	return d, nil
}

func newToken() (string, error) {
	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("make a token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
