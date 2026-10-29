package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// call is one command a fake Runner was asked to run.
type call struct{ cmd, stdin string }

// fakeRunner answers commands from a function and remembers them.
type fakeRunner struct {
	calls  []call
	answer func(c call) (Result, error)
}

func (f *fakeRunner) Run(_ context.Context, cmd string, stdin io.Reader) (Result, error) {
	var in string
	if stdin != nil {
		b, _ := io.ReadAll(stdin)
		in = string(b)
	}
	c := call{cmd: cmd, stdin: in}
	f.calls = append(f.calls, c)
	if f.answer != nil {
		return f.answer(c)
	}
	return Result{}, nil
}

func (f *fakeRunner) commands() []string {
	out := make([]string, len(f.calls))
	for i, c := range f.calls {
		out[i] = c.cmd
	}
	return out
}

// goodFacts is the machine of the test VM: Ubuntu 24.04 on arm64, 4 GB, 2 CPUs.
func goodFacts() map[string]string {
	return map[string]string{
		"uid": "1000", "os_id": "ubuntu", "os_version": "24.04", "os_name": "Ubuntu 24.04.5 LTS", "arch": "aarch64",
		"mem_mb": "3901", "cpus": "2", "disk_gb": "36", "systemd": "yes", "cgroup": "cgroup2fs", "sudo": "yes",
		"have_curl": "yes", "have_git": "yes", "have_tar": "yes", "have_sha256sum": "yes", "have_docker": "no", "have_ss": "yes",
		"ports": "", "installed": "", "service": "inactive",
	}
}

func status(r Report, id string) Check {
	for _, c := range r.Checks {
		if c.ID == id {
			return c
		}
	}
	return Check{}
}

func TestAHealthyServerPassesEveryCheck(t *testing.T) {
	r := Evaluate(goodFacts())
	if !r.OK() || r.Arch != "arm64" || r.Root || r.HasDocker || r.Running {
		t.Errorf("report = %+v", r)
	}
	for _, c := range r.Checks {
		if c.Status != Pass {
			t.Errorf("%s = %s (%s)", c.ID, c.Status, c.Detail)
		}
	}
}

func TestEachProblemIsFoundWithItsCauseAndHowToFixIt(t *testing.T) {
	cases := []struct {
		name   string
		change map[string]string
		id     string
		want   Status
	}{
		{"an old Ubuntu", map[string]string{"os_version": "20.04"}, "os", Warn},
		{"Debian 12", map[string]string{"os_id": "debian", "os_version": "12"}, "os", Pass},
		{"another family", map[string]string{"os_id": "fedora", "os_name": "Fedora 40"}, "os", Fail},
		{"a 32-bit ARM", map[string]string{"arch": "armv7l"}, "arch", Fail},
		{"no systemd", map[string]string{"systemd": "no"}, "systemd", Fail},
		{"cgroup v1", map[string]string{"cgroup": "tmpfs"}, "cgroup", Fail},
		{"no sudo", map[string]string{"sudo": "no"}, "sudo", Fail},
		{"low memory", map[string]string{"mem_mb": "2048"}, "ram", Warn},
		{"far too little memory", map[string]string{"mem_mb": "1024"}, "ram", Fail},
		{"one CPU", map[string]string{"cpus": "1"}, "cpu", Warn},
		{"small disk", map[string]string{"disk_gb": "15"}, "disk", Warn},
		{"full disk", map[string]string{"disk_gb": "4"}, "disk", Fail},
		{"port 80 taken", map[string]string{"ports": "0.0.0.0:80 "}, "ports", Warn},
		{"no ss", map[string]string{"have_ss": "no"}, "ports", Warn},
		{"no curl", map[string]string{"have_curl": "no"}, "curl", Fail},
		{"no git", map[string]string{"have_git": "no"}, "git", Fail},
	}
	for _, c := range cases {
		f := goodFacts()
		for k, v := range c.change {
			f[k] = v
		}
		got := status(Evaluate(f), c.id)
		if got.Status != c.want {
			t.Errorf("%s: %s = %s (%s), want %s", c.name, c.id, got.Status, got.Detail, c.want)
		}
		if got.Status != Pass && got.Fix == "" {
			t.Errorf("%s: a %s with no way to fix it", c.name, got.Status)
		}
	}
	f := goodFacts()
	f["systemd"] = "no"
	if Evaluate(f).OK() {
		t.Error("a failed check must stop the installation")
	}
	f = goodFacts()
	f["mem_mb"] = "2048"
	if !Evaluate(f).OK() {
		t.Error("a warning must not stop it")
	}
}

func TestRootInstalledVersionAndDockerAreRead(t *testing.T) {
	f := goodFacts()
	f["uid"], f["installed"], f["service"], f["have_docker"], f["arch"] = "0", "astrad 1.2.0", "active", "yes", "x86_64"
	r := Evaluate(f)
	if !r.Root || r.Installed != "1.2.0" || !r.Running || !r.HasDocker || r.Arch != "amd64" {
		t.Errorf("report = %+v", r)
	}
	if status(r, "sudo").Detail != "root" {
		t.Errorf("sudo detail = %q", status(r, "sudo").Detail)
	}
}

func TestProbeRunsOneScriptAndReadsItsAnswer(t *testing.T) {
	var out strings.Builder
	for k, v := range goodFacts() {
		fmt.Fprintf(&out, "%s=%s\n", k, v)
	}
	r := &fakeRunner{answer: func(call) (Result, error) { return Result{Stdout: []byte(out.String())}, nil }}
	report, err := Probe(t.Context(), r)
	if err != nil || !report.OK() || len(r.calls) != 1 || r.calls[0].cmd != "sh -s" {
		t.Fatalf("Probe = %+v, %v, calls %v", report, err, r.commands())
	}
	for _, want := range []string{"/etc/os-release", "/proc/meminfo", "cgroup", "sudo -n true", "ss -ltnH"} {
		if !strings.Contains(r.calls[0].stdin, want) {
			t.Errorf("the probe does not look at %s", want)
		}
	}
	bad := &fakeRunner{answer: func(call) (Result, error) { return Result{Exit: 1, Stderr: []byte("boom")}, nil }}
	if _, err := Probe(t.Context(), bad); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Errorf("a failed probe: %v", err)
	}
	down := &fakeRunner{answer: func(call) (Result, error) { return Result{}, errors.New("connection lost") }}
	if _, err := Probe(t.Context(), down); err == nil {
		t.Error("a lost connection must be an error")
	}
}

// binaries makes a folder of fake binaries for arch, with a matching SHA256SUMS if asked.
func binaries(t *testing.T, arch string, sums, web bool) string {
	t.Helper()
	dir := t.TempDir()
	for _, n := range []string{"astrad", "astra-priv"} {
		if err := os.WriteFile(filepath.Join(dir, n+"-linux-"+arch), []byte("binary of "+n), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if sums {
		a, _ := fileSum(filepath.Join(dir, "astrad-linux-"+arch))
		p, _ := fileSum(filepath.Join(dir, "astra-priv-linux-"+arch))
		list := fmt.Sprintf("%s  astrad-linux-%s\n%s  astra-priv-linux-%s\n", a, arch, p, arch)
		if err := os.WriteFile(filepath.Join(dir, "SHA256SUMS"), []byte(list), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if web {
		if err := os.MkdirAll(filepath.Join(dir, "web", "assets"), 0o755); err != nil {
			t.Fatal(err)
		}
		_ = os.WriteFile(filepath.Join(dir, "web", "index.html"), []byte("<html>"), 0o644)
		_ = os.WriteFile(filepath.Join(dir, "web", "assets", "app.js"), []byte("js"), 0o644)
	}
	return dir
}

func TestAPlanIsRefusedWhenSomethingFailedOrIsNotThere(t *testing.T) {
	good := Evaluate(goodFacts())
	failing := goodFacts()
	failing["systemd"] = "no"
	arm := binaries(t, "arm64", false, false)
	for name, c := range map[string]struct {
		rep Report
		src Source
	}{
		"a failed check":     {Evaluate(failing), Source{BinaryDir: arm}},
		"no source":          {good, Source{}},
		"two sources":        {good, Source{BinaryDir: arm, ReleaseURL: "https://x.example/r"}},
		"a URL with a quote": {good, Source{ReleaseURL: "https://x.example/r'; rm -rf /"}},
		"a URL with a query": {good, Source{ReleaseURL: "https://x.example/r?x=1"}},
		"plain http":         {good, Source{ReleaseURL: "http://x.example/r"}},
		"binaries for amd64": {good, Source{BinaryDir: binaries(t, "amd64", false, false)}},
		"a missing folder":   {good, Source{BinaryDir: filepath.Join(t.TempDir(), "none")}},
	} {
		if _, err := NewPlan(&fakeRunner{}, c.rep, Options{Source: c.src}); err == nil {
			t.Errorf("%s: the plan was made", name)
		}
	}
	if _, err := NewPlan(&fakeRunner{}, Report{Checks: nil}, Options{Source: Source{BinaryDir: arm}}); err == nil {
		t.Error("a machine of no known architecture was planned for")
	}
}

func TestBinariesThatDoNotMatchTheirChecksumsAreRefusedBeforeAnythingIsSent(t *testing.T) {
	dir := binaries(t, "arm64", true, false)
	if err := os.WriteFile(filepath.Join(dir, "astrad-linux-arm64"), []byte("tampered"), 0o755); err != nil {
		t.Fatal(err)
	}
	r := &fakeRunner{}
	_, err := NewPlan(r, Evaluate(goodFacts()), Options{Source: Source{BinaryDir: dir}})
	if err == nil || !strings.Contains(err.Error(), "does not match SHA256SUMS") {
		t.Fatalf("err = %v", err)
	}
	if len(r.calls) != 0 {
		t.Error("a command ran during planning")
	}
}

func TestThePlanSaysWhatWillHappenAndWhereTheChecksumsCameFrom(t *testing.T) {
	rep := Evaluate(goodFacts())
	withSums, err := NewPlan(&fakeRunner{}, rep, Options{Source: Source{BinaryDir: binaries(t, "arm64", true, false)}})
	if err != nil || !strings.Contains(strings.Join(withSums.Notes, "\n"), "match the SHA256SUMS file") {
		t.Fatalf("notes = %v, %v", withSums.Notes, err)
	}
	plain, _ := NewPlan(&fakeRunner{}, rep, Options{Source: Source{BinaryDir: binaries(t, "arm64", false, false)}})
	if !strings.Contains(strings.Join(plain.Notes, "\n"), "only prove the upload was not damaged") {
		t.Errorf("notes = %v", plain.Notes)
	}
	titles := strings.Join(plain.Titles(), "\n")
	for _, want := range []string{"Upload the daemon", "Verify the checksums on the server", "Install astrad and astra-priv", "Create the astrad user",
		"Let astrad run astra-priv as root, and nothing else", "Install the astrad service", "Start the daemon", "Wait for the daemon"} {
		if !strings.Contains(titles, want) {
			t.Errorf("no step %q in:\n%s", want, titles)
		}
	}
	for _, not := range []string{"Docker", "web UI"} {
		if strings.Contains(titles, not) {
			t.Errorf("a step about %s that was not asked for", not)
		}
	}

	upgrade := goodFacts()
	upgrade["installed"], upgrade["service"] = "astrad 0.9.0", "active"
	withDocker, _ := NewPlan(&fakeRunner{}, Evaluate(upgrade), Options{Source: Source{BinaryDir: binaries(t, "arm64", false, true)}, Docker: true})
	all := strings.Join(withDocker.Titles(), "\n")
	if !strings.Contains(all, "Docker Engine") || !strings.Contains(all, "web UI") {
		t.Errorf("steps = %s", all)
	}
	if !strings.Contains(strings.Join(withDocker.Notes, "\n"), "astrad 0.9.0 is installed already") {
		t.Errorf("notes = %v", withDocker.Notes)
	}
	hasDocker := goodFacts()
	hasDocker["have_docker"] = "yes"
	skipped, _ := NewPlan(&fakeRunner{}, Evaluate(hasDocker), Options{Source: Source{BinaryDir: binaries(t, "arm64", false, false)}, Docker: true})
	if strings.Contains(strings.Join(skipped.Titles(), "\n"), "Docker Engine") {
		t.Error("Docker was installed again")
	}
}

// installRunner plays a server that does everything right.
func installRunner(version string) *fakeRunner {
	return &fakeRunner{answer: func(c call) (Result, error) {
		switch {
		case strings.HasPrefix(c.cmd, "mktemp"):
			return Result{Stdout: []byte("/tmp/astra-install.AbC123\n")}, nil
		case strings.Contains(c.stdin, "/api/v1/version"):
			return Result{Stdout: []byte(version + "\n")}, nil
		}
		return Result{}, nil
	}}
}

func TestInstallingUploadsVerifiesInstallsAndStartsInOrder(t *testing.T) {
	r := installRunner(`{"version":"1.0.0","api":1}`)
	plan, err := NewPlan(r, Evaluate(goodFacts()), Options{Source: Source{BinaryDir: binaries(t, "arm64", false, true)}})
	if err != nil {
		t.Fatal(err)
	}
	var seen []string
	out, err := plan.Run(t.Context(), func(_ int, title string) { seen = append(seen, title) })
	if err != nil || out.Version != `{"version":"1.0.0","api":1}` {
		t.Fatalf("Run = %+v, %v", out, err)
	}
	if len(seen) != len(plan.Steps) {
		t.Errorf("progress told about %d of %d steps", len(seen), len(plan.Steps))
	}
	dir := "/tmp/astra-install.AbC123"
	want := []string{
		"mktemp -d /tmp/astra-install.XXXXXX",
		"cat > " + dir + "/astrad && chmod 0755 " + dir + "/astrad",
		"cat > " + dir + "/astra-priv && chmod 0755 " + dir + "/astra-priv",
		"cat > " + dir + "/SHA256SUMS",
		"mkdir " + dir + "/web && tar -xzf - -C " + dir + "/web",
		"sh -s",                                           // verify
		"sudo -n sh -s", "sudo -n sh -s", "sudo -n sh -s", // binaries, user, sudoers
		"sudo -n sh -s", // web
		"sudo -n sh -s", // service
		"sudo -n sh -s", // start
		"sh -s",         // wait
		"rm -rf " + dir,
	}
	if got := r.commands(); strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("commands:\n got %q\nwant %q", got, want)
	}
	if r.calls[1].stdin != "binary of astrad" || r.calls[2].stdin != "binary of astra-priv" {
		t.Errorf("uploaded %q and %q", r.calls[1].stdin, r.calls[2].stdin)
	}
	if !strings.Contains(r.calls[3].stdin, "  astrad\n") || !strings.Contains(r.calls[3].stdin, "  astra-priv\n") {
		t.Errorf("checksums sent: %q", r.calls[3].stdin)
	}
	if !strings.Contains(r.calls[5].stdin, "sha256sum -c SHA256SUMS") {
		t.Errorf("verify script: %q", r.calls[5].stdin)
	}
	if s := r.calls[6].stdin; !strings.Contains(s, dir+"/astrad") || strings.Contains(s, "@DIR@") || !strings.HasPrefix(s, "set -eu\n") {
		t.Errorf("install script: %q", s)
	}
}

// unitOf installs on a fake server and returns the service unit it wrote.
func unitOf(t *testing.T, web bool) string {
	t.Helper()
	r := installRunner("{}")
	plan, _ := NewPlan(r, Evaluate(goodFacts()), Options{Source: Source{BinaryDir: binaries(t, "arm64", false, web)}})
	if _, err := plan.Run(t.Context(), nil); err != nil {
		t.Fatal(err)
	}
	for _, c := range r.calls {
		if strings.Contains(c.stdin, "/etc/systemd/system/astrad.service") {
			return c.stdin
		}
	}
	t.Fatal("no unit was written")
	return ""
}

func TestAServiceUnitNamesItsDatabaseAndRestartsOnFailure(t *testing.T) {
	unit := unitOf(t, false)
	for _, want := range []string{"User=astrad", "Environment=ASTRA_DB=/var/lib/astrad/astra.db", "ExecStart=/usr/local/bin/astrad serve", "Restart=on-failure", "WantedBy=multi-user.target", "daemon-reload"} {
		if !strings.Contains(unit, want) {
			t.Errorf("the unit lacks %q", want)
		}
	}
	if strings.Contains(unit, "NoNewPrivileges") {
		t.Error("the daemon needs sudo, so NoNewPrivileges would break it")
	}
}

func TestTheUnitServesTheWebUIOnlyWhenThereIsOne(t *testing.T) {
	for _, web := range []bool{true, false} {
		if got := strings.Contains(unitOf(t, web), "Environment=ASTRA_WEB_DIR=/var/lib/astra/web"); got != web {
			t.Errorf("web=%v: ASTRA_WEB_DIR in the unit = %v", web, got)
		}
	}
}

func TestInstallingAsRootNeedsNoSudo(t *testing.T) {
	f := goodFacts()
	f["uid"] = "0"
	r := installRunner("{}")
	plan, _ := NewPlan(r, Evaluate(f), Options{Source: Source{BinaryDir: binaries(t, "arm64", false, false)}})
	if _, err := plan.Run(t.Context(), nil); err != nil {
		t.Fatal(err)
	}
	for _, c := range r.commands() {
		if strings.Contains(c, "sudo") {
			t.Errorf("sudo used as root: %q", c)
		}
	}
}

func TestTheStepsThatTouchTheSystemAreTheOnesTheDesignSays(t *testing.T) {
	for name, want := range map[string][]string{
		installBinaries: {".previous", "install -m 0755 -o root -g root", "mv -f"},
		createUser:      {"groupadd --system astrad", "--shell /usr/sbin/nologin", "install -d -o astrad -g astrad -m 0700 /var/lib/astrad", "/var/lib/astra/agents"},
		allowHelper:     {"astrad ALL=(root) NOPASSWD: /usr/libexec/astra/astra-priv", "visudo -cf", "chmod 0440"},
		startDaemon:     {"systemctl enable astrad", "systemctl restart astrad"},
		installDocker:   {"apt-get install -y -qq docker.io", "systemctl enable --now docker"},
	} {
		for _, w := range want {
			if !strings.Contains(name, w) {
				t.Errorf("a script lacks %q:\n%s", w, name)
			}
		}
	}
	if strings.Contains(allowHelper, "ALL=(ALL)") {
		t.Error("the sudo rule must name one program")
	}
}

func TestInstallStopsAtTheFirstFailureNamesItAndCleansUp(t *testing.T) {
	r := installRunner("{}")
	base := r.answer
	r.answer = func(c call) (Result, error) {
		if strings.Contains(c.stdin, "groupadd") {
			return Result{Exit: 9, Stderr: []byte("useradd: cannot lock /etc/passwd")}, nil
		}
		return base(c)
	}
	plan, _ := NewPlan(r, Evaluate(goodFacts()), Options{Source: Source{BinaryDir: binaries(t, "arm64", false, false)}})
	_, err := plan.Run(t.Context(), nil)
	if err == nil || !strings.Contains(err.Error(), "Create the astrad user") || !strings.Contains(err.Error(), "cannot lock /etc/passwd") {
		t.Fatalf("err = %v", err)
	}
	cmds := r.commands()
	if last := cmds[len(cmds)-1]; last != "rm -rf /tmp/astra-install.AbC123" {
		t.Errorf("the temporary folder was not removed: last command %q", last)
	}
	for _, c := range cmds {
		if strings.Contains(c, "systemctl") {
			t.Error("a later step ran after the failure")
		}
	}
	for _, c := range r.calls {
		if strings.Contains(c.stdin, "systemctl restart") {
			t.Error("the daemon was started after a failed step")
		}
	}
}

func TestFailedChecksumsInstallNothing(t *testing.T) {
	r := installRunner("{}")
	base := r.answer
	r.answer = func(c call) (Result, error) {
		if strings.Contains(c.stdin, "sha256sum -c") {
			return Result{Exit: 1, Stderr: []byte("astrad: FAILED")}, nil
		}
		return base(c)
	}
	plan, _ := NewPlan(r, Evaluate(goodFacts()), Options{Source: Source{BinaryDir: binaries(t, "arm64", false, false)}})
	_, err := plan.Run(t.Context(), nil)
	if err == nil || !strings.Contains(err.Error(), "nothing was installed") {
		t.Fatalf("err = %v", err)
	}
	for _, c := range r.calls {
		if strings.HasPrefix(c.cmd, "sudo") {
			t.Error("a privileged step ran after the checksums failed")
		}
	}
}

func TestAServerThatNamesAnUnexpectedTempFolderIsNotTrusted(t *testing.T) {
	for _, bad := range []string{"/etc", "/tmp/astra-install.x; rm -rf /", "", "/tmp/astra-install.../../etc"} {
		r := &fakeRunner{answer: func(call) (Result, error) { return Result{Stdout: []byte(bad)}, nil }}
		plan, _ := NewPlan(r, Evaluate(goodFacts()), Options{Source: Source{BinaryDir: binaries(t, "arm64", false, false)}})
		if _, err := plan.Run(t.Context(), nil); err == nil {
			t.Errorf("temp folder %q was accepted", bad)
		}
		for _, c := range r.commands() {
			if strings.HasPrefix(c, "rm -rf") {
				t.Errorf("removed %q", c)
			}
		}
	}
}

func TestAReleaseIsDownloadedAndVerifiedOnTheServer(t *testing.T) {
	r := installRunner("{}")
	rep := Evaluate(goodFacts())
	plan, err := NewPlan(r, rep, Options{Source: Source{ReleaseURL: "https://releases.example.com/astra/v1.0.0/"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := plan.Run(t.Context(), nil); err != nil {
		t.Fatal(err)
	}
	var download, verify string
	for _, c := range r.calls {
		switch {
		case strings.Contains(c.stdin, "curl -fsSL"):
			download = c.stdin
		case strings.Contains(c.stdin, "tar -xzf"):
			verify = c.stdin
		}
	}
	for _, want := range []string{"https://releases.example.com/astra/v1.0.0/SHA256SUMS", "https://releases.example.com/astra/v1.0.0/astra-linux-arm64.tar.gz"} {
		if !strings.Contains(download, want) {
			t.Errorf("the download script lacks %q:\n%s", want, download)
		}
	}
	if !strings.Contains(verify, `grep -E '  astra-linux-arm64\.tar\.gz$' SHA256SUMS | sha256sum -c -`) {
		t.Errorf("verify script:\n%s", verify)
	}
	if !strings.Contains(strings.Join(plan.Titles(), "\n"), "web UI") {
		t.Error("a release carries the web UI")
	}
	mirror, err := NewPlan(r, rep, Options{Source: Source{ReleaseURL: "http://10.0.2.2:8000/r", AllowHTTP: true}})
	if err != nil || len(mirror.Steps) == 0 {
		t.Errorf("an http mirror with opt-in: %v", err)
	}
}

func TestPairingMakesATokenAsTheDaemonUserAndTradesItWithoutShowingIt(t *testing.T) {
	const pairing = "abcdefghijklmnopqrstuvwxyz0123456789_-ABCDE"
	for _, root := range []bool{false, true} {
		r := &fakeRunner{answer: func(call) (Result, error) { return Result{Stdout: []byte(pairing + "\n")}, nil }}
		var gotPairing, gotName string
		token, err := Pair(t.Context(), r, root, "Maulana's laptop", func(_ context.Context, p, name string) (string, error) {
			gotPairing, gotName = p, name
			return "device-token-value", nil
		})
		if err != nil || token.Reveal() != "device-token-value" || gotPairing != pairing || gotName != "Maulana's laptop" {
			t.Fatalf("root=%v: %q, %v, traded %q %q", root, token.Reveal(), err, gotPairing, gotName)
		}
		want := pairAsSudoer
		if root {
			want = pairAsRoot
		}
		if r.calls[0].cmd != want || !strings.Contains(want, "ASTRA_DB=/var/lib/astrad/astra.db") {
			t.Errorf("command = %q", r.calls[0].cmd)
		}
		if strings.Contains(fmt.Sprint(token), "device-token-value") || strings.Contains(fmt.Sprintf("%v %+v", token, token), pairing) {
			t.Error("a token printed its value")
		}
	}
}

func TestPairingFailuresNeverCarryATokenAndSayWhatToDo(t *testing.T) {
	const pairing = "abcdefghijklmnopqrstuvwxyz0123456789_-ABCDE"
	ok := func(call) (Result, error) { return Result{Stdout: []byte(pairing)}, nil }
	exchange := func(context.Context, string, string) (string, error) { return "device", nil }

	notInstalled := &fakeRunner{answer: func(call) (Result, error) { return Result{Exit: 127, Stderr: []byte("astrad: not found")}, nil }}
	if _, err := Pair(t.Context(), notInstalled, false, "x", exchange); err == nil || !strings.Contains(err.Error(), "run the install first") {
		t.Errorf("not installed: %v", err)
	}
	noise := &fakeRunner{answer: func(call) (Result, error) { return Result{Stdout: []byte("Welcome to Ubuntu\nplease wait")}, nil }}
	if _, err := Pair(t.Context(), noise, false, "x", exchange); err == nil || !strings.Contains(err.Error(), "did not print a pairing token") {
		t.Errorf("noise instead of a token: %v", err)
	}
	_, err := Pair(t.Context(), &fakeRunner{answer: ok}, false, "x", func(context.Context, string, string) (string, error) {
		return "", errors.New("the daemon answered 401")
	})
	if err == nil || strings.Contains(err.Error(), pairing) || !strings.Contains(err.Error(), "401") {
		t.Errorf("exchange failure: %v", err)
	}
	if _, err := Pair(t.Context(), &fakeRunner{answer: ok}, false, "x", func(context.Context, string, string) (string, error) { return "", nil }); err == nil {
		t.Error("an empty device token was accepted")
	}
}

func registryWith(t *testing.T) (Registry, Entry) {
	t.Helper()
	g := Registry{Path: filepath.Join(t.TempDir(), "astra", "servers.json")}
	vps := Entry{Name: "vps", Host: "203.0.113.7", Port: 22, User: "dev", KeyPath: "/home/me/.config/astra/keys/astra", Role: "all-in-one"}
	if err := g.Save(vps); err != nil {
		t.Fatal(err)
	}
	if err := g.Save(Entry{Name: "second", Host: "second.example.com", Port: 2222, User: "root", KeyPath: "k"}); err != nil {
		t.Fatal(err)
	}
	return g, vps
}

func TestTheRegistryKeepsServersInAFileOnlyTheOwnerCanRead(t *testing.T) {
	g, vps := registryWith(t)
	vps.DeviceToken = "paired"
	if err := g.Save(vps); err != nil {
		t.Fatal(err)
	}
	list, _ := g.List()
	if len(list) != 2 || list[0].Name != "vps" || !list[0].Paired() || list[1].Paired() {
		t.Errorf("list = %+v", list)
	}
	if info, _ := os.Stat(g.Path); info.Mode().Perm() != 0o600 {
		t.Errorf("mode = %v", info.Mode().Perm())
	}
	if got, err := g.Get("second"); err != nil || got.Port != 2222 {
		t.Errorf("Get = %+v, %v", got, err)
	}
	if _, err := g.Get("nope"); !errors.Is(err, ErrNotFound) {
		t.Errorf("an unknown server: %v", err)
	}
}

func TestServersCanBeRemovedAndABrokenFileIsNotAnEmptyList(t *testing.T) {
	g, _ := registryWith(t)
	if list, err := (Registry{Path: filepath.Join(t.TempDir(), "none.json")}).List(); err != nil || len(list) != 0 {
		t.Fatalf("an empty registry: %v, %v", list, err)
	}
	if err := g.Remove("vps"); err != nil {
		t.Fatal(err)
	}
	if err := g.Remove("vps"); !errors.Is(err, ErrNotFound) {
		t.Errorf("removing twice: %v", err)
	}
	if list, _ := g.List(); len(list) != 1 {
		t.Errorf("after removing: %+v", list)
	}
	if err := os.WriteFile(g.Path, []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := g.List(); err == nil {
		t.Error("a broken file must be an error, not an empty list")
	}
}

func TestServersAreValidatedBeforeTheyReachACommandOrAFile(t *testing.T) {
	ok := Entry{Name: "vps", Host: "203.0.113.7", Port: 22, User: "dev", KeyPath: "k"}
	if err := ok.Validate(); err != nil {
		t.Fatal(err)
	}
	for name, change := range map[string]func(*Entry){
		"an uppercase name":        func(e *Entry) { e.Name = "VPS" },
		"a name with a slash":      func(e *Entry) { e.Name = "a/b" },
		"an empty name":            func(e *Entry) { e.Name = "" },
		"a host with a space":      func(e *Entry) { e.Host = "a b" },
		"a host with a quote":      func(e *Entry) { e.Host = "a';id" },
		"a host with a dash first": func(e *Entry) { e.Host = "-oProxyCommand=x" },
		"a user with a dollar":     func(e *Entry) { e.User = "a$b" },
		"port 0":                   func(e *Entry) { e.Port = 0 },
		"port 70000":               func(e *Entry) { e.Port = 70000 },
		"no key":                   func(e *Entry) { e.KeyPath = "" },
	} {
		e := ok
		change(&e)
		if err := e.Validate(); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
	g := Registry{Path: filepath.Join(t.TempDir(), "s.json")}
	bad := ok
	bad.Name = "A"
	if err := g.Save(bad); err == nil {
		t.Error("an invalid server was saved")
	}
}

func TestTheWebUIIsPackedForTheUploadWithoutLinksOrOwners(t *testing.T) {
	dir := binaries(t, "arm64", false, true)
	if err := os.Symlink("/etc/passwd", filepath.Join(dir, "web", "link")); err != nil {
		t.Fatal(err)
	}
	r, err := tarWeb(filepath.Join(dir, "web"))
	if err != nil {
		t.Fatal(err)
	}
	data, _ := io.ReadAll(r)
	if len(data) == 0 {
		t.Fatal("empty archive")
	}
	if strings.Contains(string(data), "passwd") {
		t.Error("a symbolic link was packed")
	}
}

func TestTheTemporaryFolderIsRemovedWhenADownloadFails(t *testing.T) {
	r := installRunner("{}")
	base := r.answer
	r.answer = func(c call) (Result, error) {
		if strings.Contains(c.stdin, "curl -fsSL") {
			return Result{Exit: 22, Stderr: []byte("curl: (22) The requested URL returned error: 404")}, nil
		}
		return base(c)
	}
	plan, err := NewPlan(r, Evaluate(goodFacts()), Options{Source: Source{ReleaseURL: "https://releases.astra-app.id/v1"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := plan.Run(t.Context(), nil); err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("err = %v", err)
	}
	cmds := r.commands()
	if last := cmds[len(cmds)-1]; last != "rm -rf /tmp/astra-install.AbC123" {
		t.Errorf("the temporary folder was not removed: last command %q", last)
	}
}

func TestACancelledInstallStillCleansUp(t *testing.T) {
	r := installRunner("{}")
	plan, _ := NewPlan(r, Evaluate(goodFacts()), Options{Source: Source{BinaryDir: binaries(t, "arm64", false, false)}})
	ctx, cancel := context.WithCancel(t.Context())
	base := r.answer
	r.answer = func(c call) (Result, error) {
		if strings.HasSuffix(c.cmd, "/astrad && chmod 0755 /tmp/astra-install.AbC123/astrad") {
			cancel() // the person pressed Ctrl-C during the upload
			return Result{}, ctx.Err()
		}
		return base(c)
	}
	if _, err := plan.Run(ctx, nil); err == nil {
		t.Fatal("a cancelled install reported success")
	}
	cmds := r.commands()
	if last := cmds[len(cmds)-1]; last != "rm -rf /tmp/astra-install.AbC123" {
		t.Errorf("after a cancel the last command was %q", last)
	}
}
