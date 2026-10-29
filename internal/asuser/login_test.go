package asuser

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalAgentLoginLooksOnlyForTheLoginFiles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	r := Runner{Root: t.TempDir(), Local: true}
	status, err := r.AgentLogin(context.Background(), "acme")
	if err != nil || status["claude"] || status["codex"] {
		t.Fatalf("nothing signed in: %v, %v", status, err)
	}
	if err := os.MkdirAll(filepath.Join(home, ".codex"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".codex", "auth.json"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	status, _ = r.AgentLogin(context.Background(), "acme")
	if status["claude"] || !status["codex"] {
		t.Errorf("status = %v", status)
	}
	if _, err := r.AgentLogin(context.Background(), "../x"); err == nil {
		t.Error("a bad slug must be refused")
	}
}
