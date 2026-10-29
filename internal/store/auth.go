package store

import (
	"context"
	"fmt"
	"time"

	"github.com/alfarax9/astra/internal/domain"
)

// CreatePairing keeps the hash of a pairing token until expires.
func (s *Store) CreatePairing(ctx context.Context, hash string, expires time.Time) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO pairings (token_hash, expires_at) VALUES (?, ?)`, hash, expires.UnixMilli())
	return wrap("create pairing", err)
}

// ConsumePairing uses up a pairing token: it succeeds once, and only before it expires. A
// token that is unknown, expired or already used gives domain.ErrNotFound. Expired tokens
// are deleted on the way.
func (s *Store) ConsumePairing(ctx context.Context, hash string, now time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return wrap("begin pairing", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM pairings WHERE expires_at <= ?`, now.UnixMilli()); err != nil {
		return wrap("purge pairings", err)
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM pairings WHERE token_hash = ?`, hash)
	if err != nil {
		return wrap("use pairing", err)
	}
	if n, err := res.RowsAffected(); err != nil || n == 0 {
		return fmt.Errorf("use pairing: %w", domain.ErrNotFound)
	}
	return wrap("use pairing", tx.Commit())
}

// CreateDevice stores a paired device with the hash of its token.
func (s *Store) CreateDevice(ctx context.Context, d domain.Device, hash string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO devices (id, name, token_hash, created_at, last_seen) VALUES (?, ?, ?, ?, ?)`,
		d.ID, d.Name, hash, d.CreatedAt.UnixMilli(), d.LastSeen.UnixMilli())
	return wrap("create device "+d.ID, err)
}

// DeviceByToken finds the device a token hash belongs to; domain.ErrNotFound if none.
func (s *Store) DeviceByToken(ctx context.Context, hash string) (domain.Device, error) {
	return scanDevice(s.db.QueryRowContext(ctx,
		`SELECT id, name, created_at, last_seen FROM devices WHERE token_hash = ?`, hash))
}

func scanDevice(r scanner) (domain.Device, error) {
	var d domain.Device
	var created, seen int64
	err := r.Scan(&d.ID, &d.Name, &created, &seen)
	d.CreatedAt = domain.Timestamp{Time: time.UnixMilli(created)}
	d.LastSeen = domain.Timestamp{Time: time.UnixMilli(seen)}
	return d, wrap("find device", err)
}

// TouchDevice records that a device was just seen.
func (s *Store) TouchDevice(ctx context.Context, id string, at time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE devices SET last_seen = ? WHERE id = ?`, at.UnixMilli(), id)
	return wrap("touch device "+id, err)
}

// Devices lists the paired devices, oldest first.
func (s *Store) Devices(ctx context.Context) ([]domain.Device, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, created_at, last_seen FROM devices ORDER BY created_at, id`)
	if err != nil {
		return nil, wrap("list devices", err)
	}
	defer func() { _ = rows.Close() }()
	var out []domain.Device
	for rows.Next() {
		d, err := scanDevice(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, wrap("list devices", rows.Err())
}

// RevokeDevice removes a device, so its token stops working at once.
func (s *Store) RevokeDevice(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM devices WHERE id = ?`, id)
	if err != nil {
		return wrap("revoke device "+id, err)
	}
	if n, err := res.RowsAffected(); err != nil || n == 0 {
		return fmt.Errorf("revoke device %s: %w", id, domain.ErrNotFound)
	}
	return nil
}
