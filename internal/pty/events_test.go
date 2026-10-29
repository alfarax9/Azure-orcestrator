package pty

import (
	"context"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

// testFollowers are the event sources the test binary's host knows, as astrad registers the
// agents' transcript readers.
var testFollowers = map[string]Follower{
	"two": func(ctx context.Context, dir string, emit func([]byte)) error {
		emit([]byte("one in " + dir))
		emit([]byte("two"))
		<-ctx.Done()
		return nil
	},
}

// records collects event records until it has n, or the deadline passes.
func records(t *testing.T, c *Conn, n int) []string {
	t.Helper()
	_ = c.SetReadDeadline(time.Now().Add(10 * time.Second))
	var got []string
	for len(got) < n {
		ev, err := c.Next()
		if err != nil {
			t.Fatalf("waiting for %d records, got %q: %v", n, got, err)
		}
		if ev.Record != nil {
			got = append(got, string(ev.Record))
		}
	}
	return got
}

func TestEventRecordsReachTheDaemonAndAreReplayedToALaterOne(t *testing.T) {
	ctx := t.Context()
	sp := spawner(t)
	spec := shell("ev1", "sleep 30")
	spec.Follow = "two"
	sock, err := sp.Start(ctx, spec)
	if err != nil {
		t.Fatal(err)
	}
	first, err := Dial(ctx, sock)
	if err != nil {
		t.Fatal(err)
	}
	if got := records(t, first, 2); got[0] != "one in /" || got[1] != "two" {
		t.Errorf("first daemon got %q", got)
	}
	_ = first.Close()

	// A restarted daemon attaches later and still gets the history, after the repaint.
	second, err := Dial(ctx, sock)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = second.Close() }()
	if got := records(t, second, 2); got[0] != "one in /" || got[1] != "two" {
		t.Errorf("second daemon got %q", got)
	}
	_ = second.Stop()
}

func TestAHostRefusesAnEventSourceItDoesNotKnow(t *testing.T) {
	sp := spawner(t)
	spec := shell("ev2", "sleep 30")
	spec.Follow = "nothing"
	if _, err := sp.Start(t.Context(), spec); err == nil {
		t.Fatal("a host with an unregistered event source must not start")
	}
	spec.Follow = "Bad;name"
	if _, err := sp.Start(t.Context(), spec); err == nil || !strings.Contains(err.Error(), "invalid event source") {
		t.Errorf("err = %v, want an invalid event source", err)
	}
}

func TestEventHistoryIsBounded(t *testing.T) {
	h := &host{attached: map[net.Conn]struct{}{}}
	for i := range maxEvents + 100 {
		h.event([]byte(strconv.Itoa(i)))
	}
	if len(h.events) != maxEvents || string(h.events[0]) != "100" {
		t.Errorf("kept %d records starting at %q", len(h.events), h.events[0])
	}
	big := make([]byte, maxEventBytes/2+1)
	h.event(big)
	h.event(big)
	if h.eventSize > maxEventBytes {
		t.Errorf("history holds %d bytes, over the %d limit", h.eventSize, maxEventBytes)
	}
}
