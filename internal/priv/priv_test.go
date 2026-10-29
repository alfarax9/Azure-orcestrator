package priv

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// fakeHost records what the helper would do to the machine.
type fakeHost struct {
	users    map[string]Agent
	groups   map[string]bool
	execs    [][]string
	dirs     []Dir
	sudoers  map[string]string
	files    map[string]bool
	execFail bool
}

func newFakeHost() *fakeHost {
	return &fakeHost{users: map[string]Agent{}, groups: map[string]bool{"astrad": true}, sudoers: map[string]string{}, files: map[string]bool{}}
}

func (f *fakeHost) Exec(_ context.Context, argv ...string) ([]byte, error) {
	f.execs = append(f.execs, argv)
	if f.execFail {
		return []byte("it broke"), errors.New("exit status 1")
	}
	return nil, nil
}

func (f *fakeHost) LookupAgent(name string) (Agent, error) {
	a, ok := f.users[name]
	if !ok {
		return Agent{}, fmt.Errorf("user %s: %w", name, fs.ErrNotExist)
	}
	return a, nil
}

func (f *fakeHost) GroupExists(name string) bool { return f.groups[name] }
func (f *fakeHost) FileExists(path string) bool  { return f.files[path] }
func (f *fakeHost) Install(d Dir) error          { f.dirs = append(f.dirs, d); return nil }

func (f *fakeHost) WriteSudoers(_ context.Context, dir, name, content string) error {
	f.sudoers[filepath.Join(dir, name)] = content
	return nil
}

func (f *fakeHost) ResolveDir(root, rel string) (string, error) { return filepath.Join(root, rel), nil }

func run(h *fakeHost, args ...string) (string, error) {
	var out bytes.Buffer
	err := Helper{Config: DefaultConfig(), Host: h}.Run(context.Background(), args, &out)
	return strings.TrimSpace(out.String()), err
}

func TestWorkspaceAddCreatesUserFoldersAndRule(t *testing.T) {
	h := newFakeHost()
	if _, err := run(h, "workspace-add", "acme"); err != nil {
		t.Fatal(err)
	}
	wantExecs := [][]string{
		{"/usr/sbin/groupadd", "--system", "astra-agent-acme"},
		{"/usr/sbin/useradd", "--system", "--gid", "astra-agent-acme", "--home-dir",
			"/var/lib/astra/agents/acme/home", "--no-create-home", "--shell", "/bin/bash", "astra-agent-acme"},
	}
	if fmt.Sprint(h.execs) != fmt.Sprint(wantExecs) {
		t.Errorf("commands = %v, want %v", h.execs, wantExecs)
	}
	modes := map[string]string{}
	for _, d := range h.dirs {
		modes[d.Path] = fmt.Sprintf("%s:%s %v", d.Owner, d.Group, d.Mode)
	}
	for path, want := range map[string]string{
		"/var/lib/astra/agents/acme":          "astra-agent-acme:astrad -rwx--x---",
		"/var/lib/astra/agents/acme/sessions": "astra-agent-acme:astrad grwxrwx---",
		"/var/lib/astra/agents/acme/work":     "astra-agent-acme:astra-agent-acme -rwx------",
		"/var/lib/astra/agents/acme/home":     "astra-agent-acme:astra-agent-acme -rwx------",
	} {
		if modes[path] != want {
			t.Errorf("%s = %q, want %q", path, modes[path], want)
		}
	}
	rule := h.sudoers["/etc/sudoers.d/astra-agent-acme"]
	if !strings.Contains(rule, "astrad ALL=(astra-agent-acme) NOPASSWD: ALL\n") {
		t.Errorf("sudoers rule = %q", rule)
	}
}

func TestWorkspaceAddCanBeRepeated(t *testing.T) {
	h := newFakeHost()
	h.users["astra-agent-acme"] = Agent{}
	h.groups["astra-agent-acme"] = true
	if _, err := run(h, "workspace-add", "acme"); err != nil {
		t.Fatal(err)
	}
	if len(h.execs) != 0 {
		t.Errorf("an existing user must not be created again, ran %v", h.execs)
	}
	if len(h.dirs) == 0 || len(h.sudoers) != 1 {
		t.Error("folders and the sudo rule must still be put right")
	}
}

func TestWorkspaceAddKeepsExistingGroupWhenOnlyUserIsMissing(t *testing.T) {
	h := newFakeHost()
	h.groups["astra-agent-acme"] = true
	if _, err := run(h, "workspace-add", "acme"); err != nil {
		t.Fatal(err)
	}
	if len(h.execs) != 1 || h.execs[0][0] != "/usr/sbin/useradd" {
		t.Errorf("commands = %v, want only useradd", h.execs)
	}
}

func TestWorkspaceAddRefusesAUserWithExtraGroups(t *testing.T) {
	h := newFakeHost()
	h.users["astra-agent-acme"] = Agent{ExtraGroups: []string{"docker"}}
	_, err := run(h, "workspace-add", "acme")
	if err == nil || !strings.Contains(err.Error(), "docker") {
		t.Fatalf("err = %v, want a refusal naming the group", err)
	}
	if len(h.dirs) != 0 || len(h.sudoers) != 0 {
		t.Error("nothing may be set up for a user that breaks the isolation")
	}
}

func TestWorkspaceAddNeedsTheDaemonGroup(t *testing.T) {
	h := newFakeHost()
	delete(h.groups, "astrad")
	if _, err := run(h, "workspace-add", "acme"); err == nil || !strings.Contains(err.Error(), "install the daemon") {
		t.Fatalf("err = %v", err)
	}
}

func TestWorkspaceAddReportsFailedCommands(t *testing.T) {
	h := newFakeHost()
	h.execFail = true
	if _, err := run(h, "workspace-add", "acme"); err == nil || !strings.Contains(err.Error(), "it broke") {
		t.Fatalf("err = %v, want the command's output", err)
	}
}

func TestEveryBadRequestIsRefusedBeforeAnythingRuns(t *testing.T) {
	long := strings.Repeat("a", 21)
	cases := [][]string{
		nil,
		{"format-disk"},
		{"workspace-add"},
		{"workspace-add", "acme", "extra"},
		{"workspace-add", "ACME"},
		{"workspace-add", "a"},
		{"workspace-add", long},
		{"workspace-add", "x; rm -rf /"},
		{"workspace-add", "../etc"},
		{"workspace-add", "-rf"},
		{"session-start", "acme"},
		{"session-start", "-acme", "s1", "--agent", "shell"},
		{"session-start", "acme", "s1"},
		{"session-start", "acme", "s1", "--agent", "bash"},
		{"session-start", "acme", "S1", "--agent", "shell"},
		{"session-start", "acme", "s1", "--agent", "shell", "extra"},
		{"session-start", "acme", "s1", "--agent", "shell", "--env", "X=1"},
		{"session-start", "acme", "s1", "--agent", "shell", "--dir", "../other"},
		{"session-start", "acme", "s1", "--agent", "shell", "--dir", "/etc"},
		{"session-start", "acme", "s1", "--agent", "shell", "--dir", "a//b"},
		{"session-start", "acme", "s1", "--agent", "shell", "--dir", "a/./b"},
		{"session-start", "acme", "s1", "--agent", "shell", "--dir", "a b"},
		{"session-start", "acme", "s1", "--agent", "shell", "--memory-max", "100"},
		{"session-start", "acme", "s1", "--agent", "shell", "--memory-max", "99999"},
		{"session-start", "acme", "s1", "--agent", "shell", "--cpu-quota", "0"},
		{"session-start", "acme", "s1", "--agent", "shell", "--cpu-quota", "5000"},
		{"session-start", "acme", "s1", "--agent", "shell", "--rows", "1"},
		{"session-start", "acme", "s1", "--agent", "shell", "--cols", "9999"},
		{"session-stop", "acme"},
		{"session-stop", "acme", "s1;reboot"},
		{"session-stop", "Acme", "s1"},
	}
	for _, args := range cases {
		h := newFakeHost()
		h.users["astra-agent-acme"] = Agent{}
		_, err := run(h, args...)
		if !errors.Is(err, ErrUsage) {
			t.Errorf("%q: err = %v, want a usage error", args, err)
		}
		if len(h.execs)+len(h.dirs)+len(h.sudoers) != 0 {
			t.Errorf("%q: the helper changed the machine before refusing", args)
		}
	}
}

func TestSessionStartBuildsOneUnitForTheAgentUser(t *testing.T) {
	h := newFakeHost()
	h.users["astra-agent-acme"] = Agent{}
	out, err := run(h, "session-start", "acme", "0198-abc", "--agent", "claude", "--dir", "site/feat-x",
		"--rows", "30", "--cols", "100", "--memory-max", "2048", "--cpu-quota", "150")
	if err != nil {
		t.Fatal(err)
	}
	if want := "/var/lib/astra/agents/acme/sessions/0198-abc.sock"; out != want {
		t.Errorf("printed %q, want the socket %q", out, want)
	}
	if len(h.execs) != 1 {
		t.Fatalf("commands = %v, want one", h.execs)
	}
	argv := h.execs[0]
	for _, want := range []string{
		"/usr/bin/systemd-run", "--unit=astra-s-acme-0198-abc", "--slice=astra-agents-acme.slice",
		"--uid=astra-agent-acme", "--gid=astra-agent-acme",
		"--working-directory=/var/lib/astra/agents/acme/work/site/feat-x",
		"--property=MemoryMax=2048M", "--property=CPUQuota=150%", "--property=TasksMax=512",
		"--property=NoNewPrivileges=yes", "--property=ProtectProc=invisible", "--property=ProcSubset=pid",
		"--setenv=HOME=/var/lib/astra/agents/acme/home",
	} {
		if !slices.Contains(argv, want) {
			t.Errorf("missing %q in %v", want, argv)
		}
	}
	tail := argv[slices.Index(argv, "/usr/local/bin/astrad"):]
	wantTail := []string{"/usr/local/bin/astrad", "session-host", "--socket", out, "--socket-mode", "660",
		"--dir", "/var/lib/astra/agents/acme/work/site/feat-x", "--rows", "30", "--cols", "100",
		"--follow", "claude", "--", "/usr/local/bin/claude"}
	if fmt.Sprint(tail) != fmt.Sprint(wantTail) {
		t.Errorf("host command = %v, want %v", tail, wantTail)
	}
}

func TestOnlyAgentsWithATranscriptGetAnEventSource(t *testing.T) {
	h := newFakeHost()
	h.users["astra-agent-acme"] = Agent{}
	if _, err := run(h, "session-start", "acme", "s1", "--agent", "shell"); err != nil {
		t.Fatal(err)
	}
	if slices.Contains(h.execs[0], "--follow") {
		t.Errorf("a plain shell has no transcript: %v", h.execs[0])
	}
}

func TestSessionStartGivesTheAgentOnlyTheListedEnvironment(t *testing.T) {
	h := newFakeHost()
	h.users["astra-agent-acme"] = Agent{}
	if _, err := run(h, "session-start", "acme", "s1", "--agent", "shell"); err != nil {
		t.Fatal(err)
	}
	var keys []string
	for _, a := range h.execs[0] {
		if kv, ok := strings.CutPrefix(a, "--setenv="); ok {
			k, _, _ := strings.Cut(kv, "=")
			keys = append(keys, k)
		}
	}
	want := "HOME USER LOGNAME SHELL LANG PATH TERM COLORTERM"
	if got := strings.Join(keys, " "); got != want {
		t.Errorf("environment = %q, want %q", got, want)
	}
	if !slices.Contains(h.execs[0], "--working-directory=/var/lib/astra/agents/acme/work") {
		t.Error("without --dir the session starts at the root of work/")
	}
}

func TestSessionStartNeedsAWorkspace(t *testing.T) {
	h := newFakeHost()
	_, err := run(h, "session-start", "acme", "s1", "--agent", "shell")
	if err == nil || !strings.Contains(err.Error(), "workspace-add") || len(h.execs) != 0 {
		t.Fatalf("err = %v, commands = %v", err, h.execs)
	}
}

func TestSessionStartReportsAFailedUnit(t *testing.T) {
	h := newFakeHost()
	h.users["astra-agent-acme"] = Agent{}
	h.execFail = true
	if _, err := run(h, "session-start", "acme", "s1", "--agent", "shell"); err == nil || !strings.Contains(err.Error(), "it broke") {
		t.Fatalf("err = %v", err)
	}
}

func TestSessionStopStopsThatUnit(t *testing.T) {
	h := newFakeHost()
	if _, err := run(h, "session-stop", "acme", "s1"); err != nil {
		t.Fatal(err)
	}
	want := []string{"/usr/bin/systemctl", "stop", "astra-s-acme-s1.service"}
	if len(h.execs) != 1 || fmt.Sprint(h.execs[0]) != fmt.Sprint(want) {
		t.Errorf("commands = %v, want %v", h.execs, want)
	}
	h.execFail = true
	if _, err := run(h, "session-stop", "acme", "s1"); err == nil {
		t.Error("a failed stop must be reported")
	}
}

func TestResolveDirStaysInsideWork(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Join(root, "site", "feat"), 0o755))
	must(os.WriteFile(filepath.Join(root, "file"), nil, 0o644))
	must(os.Symlink(outside, filepath.Join(root, "escape")))
	must(os.Symlink("site", filepath.Join(root, "inside")))

	var h SystemHost
	real, err := filepath.EvalSymlinks(root)
	must(err)
	for rel, want := range map[string]string{"": real, "site/feat": filepath.Join(real, "site", "feat"), "inside": filepath.Join(real, "site")} {
		if got, err := h.ResolveDir(root, rel); err != nil || got != want {
			t.Errorf("ResolveDir(%q) = %q, %v, want %q", rel, got, err, want)
		}
	}
	for _, rel := range []string{"escape", "missing", "file"} {
		if got, err := h.ResolveDir(root, rel); err == nil {
			t.Errorf("ResolveDir(%q) = %q, want an error", rel, got)
		}
	}
}

func TestHostFindsGroupsAndUsers(t *testing.T) {
	var h SystemHost
	if !h.GroupExists("root") && !h.GroupExists("wheel") && !h.GroupExists("daemon") {
		t.Error("expected a system group to exist")
	}
	if h.GroupExists("astra-no-such-group") {
		t.Error("a missing group must not exist")
	}
	if _, err := h.LookupAgent("astra-no-such-user"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("err = %v, want fs.ErrNotExist", err)
	}
	out, err := h.Exec(context.Background(), "/bin/echo", "hi")
	if err != nil || strings.TrimSpace(string(out)) != "hi" {
		t.Errorf("Exec = %q, %v", out, err)
	}
}

func TestInstallSetsModeAndRefusesLinks(t *testing.T) {
	me, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	grp, err := user.LookupGroupId(me.Gid)
	if err != nil {
		t.Fatal(err)
	}
	var h SystemHost
	base := t.TempDir()
	d := Dir{Path: filepath.Join(base, "a", "work"), Owner: me.Username, Group: grp.Name, Mode: 0o700}
	for range 2 { // repeating must be fine
		if err := h.Install(d); err != nil {
			t.Fatal(err)
		}
	}
	if info, err := os.Stat(d.Path); err != nil || info.Mode().Perm() != 0o700 {
		t.Errorf("mode = %v, %v, want 0700", info, err)
	}

	// An agent could replace a folder it owns by a link; root must not follow it.
	target := t.TempDir()
	link := filepath.Join(base, "sessions")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if err := h.Install(Dir{Path: link, Owner: me.Username, Group: grp.Name, Mode: 0o777}); err == nil {
		t.Error("Install followed a symbolic link")
	}
	if info, _ := os.Stat(target); info.Mode().Perm() == 0o777 {
		t.Error("the link's target was changed")
	}
}

func TestAgentStatusSaysWhichLoginsExistWithoutReadingThem(t *testing.T) {
	h := newFakeHost()
	h.files["/var/lib/astra/agents/acme/home/.claude/.credentials.json"] = true
	out, err := run(h, "agent-status", "acme")
	if err != nil || out != `{"claude":true,"codex":false}` {
		t.Errorf("status = %q, %v", out, err)
	}
	h.files["/var/lib/astra/agents/acme/home/.codex/auth.json"] = true
	if out, _ := run(h, "agent-status", "acme"); out != `{"claude":true,"codex":true}` {
		t.Errorf("status = %q", out)
	}
	for _, args := range [][]string{{"agent-status"}, {"agent-status", "ACME"}, {"agent-status", "a", "b"}, {"agent-status", "../etc"}} {
		if _, err := run(newFakeHost(), args...); !errors.Is(err, ErrUsage) {
			t.Errorf("%q: err = %v, want a usage error", args, err)
		}
	}
}

func TestFileExistsDoesNotFollowLinksAndSeesNothingElse(t *testing.T) {
	dir := t.TempDir()
	var h SystemHost
	if h.FileExists(filepath.Join(dir, "none")) {
		t.Error("a missing file exists")
	}
	file := filepath.Join(dir, "creds")
	_ = os.WriteFile(file, []byte("secret"), 0o600)
	if !h.FileExists(file) {
		t.Error("an existing file is missing")
	}
	link := filepath.Join(dir, "link")
	_ = os.Symlink(filepath.Join(dir, "gone"), link)
	if !h.FileExists(link) {
		t.Error("a dangling link is still something at that path")
	}
}
