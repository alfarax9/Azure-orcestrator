package store

import (
	"errors"
	"testing"
	"time"

	"github.com/alfarax9/astra/internal/domain"
)

func TestAPairingTokenWorksOnceAndOnlyBeforeItExpires(t *testing.T) {
	ctx := t.Context()
	s := open(t)
	now := time.UnixMilli(1700000000000)
	if err := s.CreatePairing(ctx, "hash-a", now.Add(10*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := s.CreatePairing(ctx, "hash-b", now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := s.ConsumePairing(ctx, "hash-a", now); err != nil {
		t.Fatalf("first use: %v", err)
	}
	if err := s.ConsumePairing(ctx, "hash-a", now); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("second use: err = %v, want ErrNotFound", err)
	}
	if err := s.ConsumePairing(ctx, "hash-b", now.Add(time.Minute)); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("an expired token: err = %v, want ErrNotFound", err)
	}
	if err := s.ConsumePairing(ctx, "never-made", now); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("an unknown token: err = %v, want ErrNotFound", err)
	}
}

func TestDevicesAreFoundByTokenHashAndCanBeRevoked(t *testing.T) {
	ctx := t.Context()
	s := open(t)
	at := domain.Timestamp{Time: time.UnixMilli(1700000000000)}
	laptop := domain.Device{ID: "d1", Name: "Laptop", CreatedAt: at, LastSeen: at}
	phone := domain.Device{ID: "d2", Name: "Phone", CreatedAt: domain.Timestamp{Time: at.Add(time.Second)}, LastSeen: at}
	for d, hash := range map[domain.Device]string{laptop: "h1", phone: "h2"} {
		if err := s.CreateDevice(ctx, d, hash); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.CreateDevice(ctx, domain.Device{ID: "d3", Name: "Copy", CreatedAt: at, LastSeen: at}, "h1"); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("a token hash used twice: err = %v, want ErrConflict", err)
	}
	got, err := s.DeviceByToken(ctx, "h1")
	if err != nil || got != laptop {
		t.Fatalf("DeviceByToken = %+v, %v", got, err)
	}
	later := at.Add(time.Hour)
	if err := s.TouchDevice(ctx, "d1", later); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.DeviceByToken(ctx, "h1"); got.LastSeen.UnixMilli() != later.UnixMilli() {
		t.Errorf("last seen = %v", got.LastSeen)
	}
	list, err := s.Devices(ctx)
	if err != nil || len(list) != 2 || list[0].ID != "d1" || list[1].ID != "d2" {
		t.Errorf("Devices = %+v, %v", list, err)
	}
	if err := s.RevokeDevice(ctx, "d1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DeviceByToken(ctx, "h1"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("a revoked device: err = %v, want ErrNotFound", err)
	}
	if err := s.RevokeDevice(ctx, "d1"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("revoking twice: err = %v, want ErrNotFound", err)
	}
}

func TestEveryWriteIsReportedToSubscribersUntilTheyCancel(t *testing.T) {
	ctx := t.Context()
	s := open(t)
	var got []domain.Change
	cancel := s.Subscribe(func(c domain.Change) { got = append(got, c) })

	ws, p := seed(t, s, "acme") // a workspace and a project
	w := domain.NewWorker(p.ID, "Task", domain.AgentClaude, "astra/x", true, time.UnixMilli(1700000002000))
	if err := s.CreateWorker(ctx, w); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpdateWorker(ctx, w.ID, func(w *domain.Worker) { w.Activity = domain.ActivityIdle }); err != nil {
		t.Fatal(err)
	}
	if err := s.SetCommands(ctx, p.ID, domain.Commands{Dev: "x"}); err != nil {
		t.Fatal(err)
	}
	// Failed writes report nothing.
	_ = s.CreateWorkspace(ctx, ws)

	if len(got) != 5 {
		t.Fatalf("got %d changes, want 5: %+v", len(got), got)
	}
	if got[0].Workspace == nil || got[1].Project == nil || got[2].Worker == nil ||
		got[3].Worker == nil || got[3].Worker.Activity != domain.ActivityIdle ||
		got[4].Project == nil || got[4].Project.Commands.Dev != "x" {
		t.Errorf("changes = %+v", got)
	}
	cancel()
	if err := s.CreateWorker(ctx, domain.NewWorker(p.ID, "Another", domain.AgentCodex, "astra/y", true, time.Now())); err != nil {
		t.Fatal(err)
	}
	if len(got) != 5 {
		t.Error("a cancelled subscriber was still told")
	}
}
