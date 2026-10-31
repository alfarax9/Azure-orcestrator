package api

import (
	"encoding/json/v2"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/alfarax9/astra/internal/domain"
)

// getList reads a route that answers a JSON array.
func getList(t *testing.T, url string) (int, []map[string]any) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+goodToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	b, _ := io.ReadAll(resp.Body)
	var out []map[string]any
	_ = json.Unmarshal(b, &out)
	return resp.StatusCode, out
}

func TestUsageIsAnsweredPerDayForTheRequestedNumberOfDays(t *testing.T) {
	ts, f := newServer(t)
	f.usage = []domain.UsageDay{{Day: "2026-10-31", Agent: domain.AgentClaude, Model: "claude-sonnet-5-5", Messages: 2, InputTokens: 10, OutputTokens: 50, CacheReadTokens: 7}}
	status, got := getList(t, ts.URL+"/api/v1/usage?days=7")
	if status != 200 || len(got) != 1 || got[0]["outputTokens"] != float64(50) || got[0]["cacheReadTokens"] != float64(7) {
		t.Fatalf("usage = %d %v", status, got)
	}
	// Seven days ending today: today and the six before it, from midnight UTC.
	if want := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -6); !f.usageSince.Equal(want) {
		t.Errorf("since = %v, want %v", f.usageSince, want)
	}
}

func TestUsageWithNothingReportedIsAnEmptyListAndBadDaysAreRefused(t *testing.T) {
	ts, _ := newServer(t)
	if status, got := getList(t, ts.URL+"/api/v1/usage"); status != 200 || got == nil || len(got) != 0 {
		t.Errorf("no usage = %d %v, want 200 and []", status, got)
	}
	for _, days := range []string{"0", "-3", "x", "367"} {
		status, body := call(t, "GET", ts.URL+"/api/v1/usage?days="+days, goodToken, "")
		if status != 400 || errCode(body) != "invalid" {
			t.Errorf("days=%s = %d %v, want 400 invalid", days, status, body)
		}
	}
}
