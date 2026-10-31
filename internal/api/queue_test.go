package api

import (
	"fmt"
	"testing"

	"github.com/alfarax9/astra/internal/agent"
	"github.com/alfarax9/astra/internal/domain"
)

func TestStartingAQueuedWorkerAnswersAcceptedAndDoesNotTrackIt(t *testing.T) {
	ts, f := newServer(t)
	f.workers = []domain.Worker{{ID: "w1", Queued: true}}
	f.failWith = fmt.Errorf("worker w1: %w", agent.ErrQueued)
	status, body := call(t, "POST", ts.URL+"/api/v1/workers/w1/start", goodToken, "")
	if status != 202 {
		t.Fatalf("a queued start = %d %s, want 202", status, body)
	}
	if len(f.tracked) != 0 {
		t.Errorf("tracked %v: a worker with no session has nothing to watch", f.tracked)
	}
}
