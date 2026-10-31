package store

import (
	"context"
	"time"

	"github.com/alfarax9/astra/internal/domain"
)

// RecordUsage stores the tokens of one message of a worker. Reporting the same key again
// replaces the counts but keeps the time of the first report, so the message stays on its day.
func (s *Store) RecordUsage(ctx context.Context, id domain.WorkerID, at time.Time, u domain.Usage) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO usage (worker_id, key, at, model, input, output, cache_read, cache_write)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (worker_id, key) DO UPDATE SET model = excluded.model, input = excluded.input,
			output = excluded.output, cache_read = excluded.cache_read, cache_write = excluded.cache_write`,
		id, u.Key, at.UnixMilli(), u.Model, u.InputTokens, u.OutputTokens, u.CacheReadTokens, u.CacheWriteTokens)
	return wrap("record usage of worker "+string(id), err)
}

// UsageByDay sums the usage reported since the given time, per UTC day, agent and model,
// oldest day first.
func (s *Store) UsageByDay(ctx context.Context, since time.Time) ([]domain.UsageDay, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT strftime('%Y-%m-%d', u.at / 1000, 'unixepoch') AS day, w.agent, u.model,
			count(*), sum(u.input), sum(u.output), sum(u.cache_read), sum(u.cache_write)
		FROM usage u JOIN workers w ON w.id = u.worker_id
		WHERE u.at >= ? GROUP BY day, w.agent, u.model ORDER BY day, w.agent, u.model`, since.UnixMilli())
	if err != nil {
		return nil, wrap("sum usage", err)
	}
	defer func() { _ = rows.Close() }()
	var out []domain.UsageDay
	for rows.Next() {
		var d domain.UsageDay
		if err := rows.Scan(&d.Day, &d.Agent, &d.Model, &d.Messages, &d.InputTokens, &d.OutputTokens,
			&d.CacheReadTokens, &d.CacheWriteTokens); err != nil {
			return nil, wrap("read usage", err)
		}
		out = append(out, d)
	}
	return out, wrap("sum usage", rows.Err())
}
