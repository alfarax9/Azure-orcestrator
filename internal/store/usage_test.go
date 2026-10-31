package store

import (
	"errors"
	"testing"
	"time"

	"github.com/alfarax9/astra/internal/domain"
)

func TestUsageIsReplacedByKeyAndSummedPerDayAgentAndModel(t *testing.T) {
	s := open(t)
	ctx := t.Context()
	_, p := seed(t, s, "acme")
	claude := domain.NewWorker(p.ID, "Fix header", domain.AgentClaude, "astra/a", true, time.UnixMilli(1700000002000))
	codex := domain.NewWorker(p.ID, "Add footer", domain.AgentCodex, "astra/b", true, time.UnixMilli(1700000003000))
	for _, w := range []domain.Worker{claude, codex} {
		if err := s.CreateWorker(ctx, w); err != nil {
			t.Fatal(err)
		}
	}
	day1 := time.Date(2026, 10, 30, 23, 59, 0, 0, time.UTC)
	day2 := time.Date(2026, 10, 31, 0, 1, 0, 0, time.UTC)
	record := func(id domain.WorkerID, at time.Time, u domain.Usage) {
		t.Helper()
		if err := s.RecordUsage(ctx, id, at, u); err != nil {
			t.Fatal(err)
		}
	}
	sonnet := domain.Usage{Model: "claude-sonnet-5-5"}
	record(claude.ID, day1, domain.Usage{Key: "m1", Model: sonnet.Model, InputTokens: 100, OutputTokens: 10, CacheWriteTokens: 500})
	// The same message again, later and complete: it replaces the first report and stays on day 1.
	record(claude.ID, day2, domain.Usage{Key: "m1", Model: sonnet.Model, InputTokens: 100, OutputTokens: 40, CacheWriteTokens: 500})
	record(claude.ID, day2, domain.Usage{Key: "m2", Model: sonnet.Model, InputTokens: 8, OutputTokens: 30, CacheReadTokens: 500})
	record(codex.ID, day2, domain.Usage{Key: "session", Model: "gpt-5-codex", InputTokens: 200, OutputTokens: 340, CacheReadTokens: 1000})

	got, err := s.UsageByDay(ctx, day1.Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	want := []domain.UsageDay{
		{Day: "2026-10-30", Agent: domain.AgentClaude, Model: sonnet.Model, Messages: 1, InputTokens: 100, OutputTokens: 40, CacheWriteTokens: 500},
		{Day: "2026-10-31", Agent: domain.AgentClaude, Model: sonnet.Model, Messages: 1, InputTokens: 8, OutputTokens: 30, CacheReadTokens: 500},
		{Day: "2026-10-31", Agent: domain.AgentCodex, Model: "gpt-5-codex", Messages: 1, InputTokens: 200, OutputTokens: 340, CacheReadTokens: 1000},
	}
	if len(got) != len(want) {
		t.Fatalf("usage = %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row %d = %+v, want %+v", i, got[i], want[i])
		}
	}

	// A window that starts after day 1 leaves it out.
	if got, err := s.UsageByDay(ctx, day2.Add(-time.Minute)); err != nil || len(got) != 2 {
		t.Errorf("since day 2 = %+v, %v, want the two rows of day 2", got, err)
	}
	if err := s.RecordUsage(ctx, "nope", day2, domain.Usage{Key: "x", InputTokens: 1}); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("usage of an unknown worker = %v, want a conflict", err)
	}
}
