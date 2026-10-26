package asuser

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func localRunner(t *testing.T) Runner {
	t.Helper()
	r := Runner{Root: t.TempDir(), Local: true}
	if err := r.AddWorkspace(context.Background(), "acme"); err != nil {
		t.Fatal(err)
	}
	return r
}

func TestAddWorkspaceCreatesTheFoldersLocally(t *testing.T) {
	r := localRunner(t)
	for _, d := range []string{"work", "home"} {
		if info, err := os.Stat(filepath.Join(r.Root, "acme", d)); err != nil || info.Mode().Perm() != 0o700 {
			t.Errorf("%s: %v, %v", d, info, err)
		}
	}
	if err := r.AddWorkspace(context.Background(), "../x"); err == nil {
		t.Error("a bad slug must be refused")
	}
}

func TestRunWorksInTheFolderAndReturnsStdout(t *testing.T) {
	r := localRunner(t)
	ctx := context.Background()
	if err := os.MkdirAll(filepath.Join(r.Root, "acme", "work", "site", "repo"), 0o700); err != nil {
		t.Fatal(err)
	}
	out, err := r.Run(ctx, "acme", "site/repo", "pwd")
	want, _ := filepath.EvalSymlinks(filepath.Join(r.Root, "acme", "work", "site", "repo"))
	if got, _ := filepath.EvalSymlinks(strings.TrimSpace(string(out))); err != nil || got != want {
		t.Errorf("pwd = %q, %v, want %q", out, err, want)
	}
	if _, err := r.Run(ctx, "acme", "", "sh", "-c", "echo broken >&2; exit 3"); err == nil || !strings.Contains(err.Error(), "broken") {
		t.Errorf("err = %v, want it to carry stderr", err)
	}
}

func TestRunPassesOnlyTheListedEnvironment(t *testing.T) {
	r := localRunner(t)
	t.Setenv("ASTRA_SECRET_TOKEN", "do-not-leak")
	out, err := r.Run(context.Background(), "acme", "", "env")
	if err != nil {
		t.Fatal(err)
	}
	var keys []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		k, _, _ := strings.Cut(line, "=")
		keys = append(keys, k)
	}
	slices.Sort(keys)
	want := []string{"GIT_ALLOW_PROTOCOL", "GIT_SSH_COMMAND", "GIT_TERMINAL_PROMPT", "HOME", "LANG", "LOGNAME", "PATH", "USER"}
	if !slices.Equal(keys, want) && !slices.Equal(slices.DeleteFunc(keys, func(k string) bool { return k == "PWD" || k == "SHLVL" || k == "_" }), want) {
		t.Errorf("environment = %v, want %v", keys, want)
	}
}

func TestRunRefusesBadTargets(t *testing.T) {
	r := localRunner(t)
	for _, c := range [][2]string{{"ACME", ""}, {"acme", ".."}, {"acme", "a/../b"}, {"acme", "/etc"}, {"acme", "a b"}} {
		if _, err := r.Run(context.Background(), c[0], c[1], "true"); err == nil {
			t.Errorf("Run(%q, %q) must be refused", c[0], c[1])
		}
	}
	if _, err := r.Run(context.Background(), "acme", ""); err == nil {
		t.Error("an empty command must be refused")
	}
}

func TestRunBoundsTheOutput(t *testing.T) {
	r := localRunner(t)
	_, err := r.Run(context.Background(), "acme", "", "head", "-c", "5000000", "/dev/zero")
	if err == nil || !strings.Contains(err.Error(), "more than") {
		t.Errorf("err = %v, want the output limit", err)
	}
}
