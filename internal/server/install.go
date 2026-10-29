package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Options is what a person asked for.
type Options struct {
	Source Source
	// Docker installs Docker Engine when it is missing. It is only ever done when asked for.
	Docker bool
}

// Step is one action of the installation, with the sentence shown for approval.
type Step struct {
	Title string
	run   func(ctx context.Context) error
}

// Plan is the installation, ready to show and then to run (F1-04).
type Plan struct {
	Steps []Step
	// Notes are things to know before agreeing: where the checksums come from, what is replaced.
	Notes []string
	in    *installer
}

// Outcome is what a finished installation reports.
type Outcome struct {
	// Version is what the daemon answered at /api/v1/version.
	Version string
}

// tempDirPattern is the only shape of remote temp folder that goes into a script.
var tempDirPattern = regexp.MustCompile(`^/tmp/astra-install\.[A-Za-z0-9]{6,12}$`)

type installer struct {
	r       Runner
	rep     Report
	opts    Options
	files   localFiles
	dir     string
	version string
}

// NewPlan decides what to do on the machine the report describes. It refuses to plan on a
// machine with a failed check, and checks that the files to install are really there.
func NewPlan(r Runner, rep Report, opts Options) (Plan, error) {
	if !rep.OK() {
		return Plan{}, errors.New("the server does not pass the readiness check; fix the failed items first")
	}
	if rep.Arch == "" {
		return Plan{}, errors.New("unsupported architecture")
	}
	if err := opts.Source.validate(); err != nil {
		return Plan{}, err
	}
	in := &installer{r: r, rep: rep, opts: opts}
	var notes []string
	if opts.Source.BinaryDir != "" {
		files, err := inspect(opts.Source.BinaryDir, rep.Arch)
		if err != nil {
			return Plan{}, err
		}
		in.files = files
		if files.provided {
			notes = append(notes, "The binaries match the SHA256SUMS file next to them, and are checked again on the server.")
		} else {
			notes = append(notes, "There is no SHA256SUMS file next to the binaries: the checksums were made here, so they only prove the upload was not damaged, not who built the files.")
		}
	}
	if rep.Installed != "" {
		notes = append(notes, fmt.Sprintf("astrad %s is installed already. It is replaced, the old files are kept as .previous, and agent sessions that are running keep running.", rep.Installed))
	}
	if opts.Docker && rep.HasDocker {
		notes = append(notes, "Docker is installed already, so it is left alone.")
	}
	return Plan{Steps: in.steps(), Notes: notes, in: in}, nil
}

// Titles lists the steps in plain words, for the person to approve.
func (p Plan) Titles() []string {
	titles := make([]string, len(p.Steps))
	for i, s := range p.Steps {
		titles[i] = s.Title
	}
	return titles
}

// Run does the steps in order and stops at the first that fails, naming it. progress hears about each step
// before it starts. The temporary folder on the server is removed in every case.
func (p Plan) Run(ctx context.Context, progress func(i int, title string)) (Outcome, error) {
	defer p.in.cleanup(ctx)
	for i, s := range p.Steps {
		if progress != nil {
			progress(i, s.Title)
		}
		if err := s.run(ctx); err != nil {
			return Outcome{}, fmt.Errorf("%s: %w", s.Title, err)
		}
	}
	return Outcome{Version: p.in.version}, nil
}

func (in *installer) steps() []Step {
	src := in.opts.Source
	var steps []Step
	if src.BinaryDir != "" {
		steps = append(steps, Step{"Upload the daemon and its helper to a temporary folder on the server", in.upload})
	} else {
		steps = append(steps, Step{"Download the release and its checksums on the server", in.download})
	}
	steps = append(steps,
		Step{"Verify the checksums on the server", in.verify},
		Step{"Install astrad and astra-priv (the old files are kept as .previous)", in.privileged(installBinaries)},
		Step{"Create the astrad user and the folders it owns", in.privileged(createUser)},
		Step{"Let astrad run astra-priv as root, and nothing else", in.privileged(allowHelper)},
	)
	if in.hasWeb() {
		steps = append(steps, Step{"Install the web UI the daemon serves", in.privileged(installWeb)})
	}
	if in.opts.Docker && !in.rep.HasDocker {
		steps = append(steps, Step{"Install Docker Engine from the distribution's packages", in.privileged(installDocker)})
	}
	return append(steps,
		Step{"Install the astrad service", in.service},
		Step{"Start the daemon (agent sessions keep running)", in.privileged(startDaemon)},
		Step{"Wait for the daemon to answer", in.wait},
	)
}

func (in *installer) hasWeb() bool { return in.files.web != "" || in.opts.Source.ReleaseURL != "" }

// cleanup removes the temporary folder; a failure to do so is not worth failing an install for.
func (in *installer) cleanup(ctx context.Context) {
	if in.dir == "" || !tempDirPattern.MatchString(in.dir) {
		return
	}
	_, _ = in.r.Run(context.WithoutCancel(ctx), "rm -rf "+in.dir, nil) // also after a cancelled install
}

// tempDir makes the folder the files go through, readable by the login user only.
func (in *installer) tempDir(ctx context.Context) error {
	res, err := in.r.Run(ctx, "mktemp -d /tmp/astra-install.XXXXXX", nil)
	if err := failed("make a temporary folder", res, err); err != nil {
		return err
	}
	dir := strings.TrimSpace(string(res.Stdout))
	if !tempDirPattern.MatchString(dir) {
		return fmt.Errorf("the server named its temporary folder %q, which is not what was asked for", dir)
	}
	in.dir = dir
	return nil
}

func (in *installer) upload(ctx context.Context) error {
	if err := in.tempDir(ctx); err != nil {
		return err
	}
	for _, f := range []struct{ local, remote string }{{in.files.astrad, "astrad"}, {in.files.priv, "astra-priv"}} {
		fh, err := os.Open(f.local)
		if err != nil {
			return fmt.Errorf("read %s: %w", f.local, err)
		}
		res, err := in.r.Run(ctx, fmt.Sprintf("cat > %s/%s && chmod 0755 %s/%s", in.dir, f.remote, in.dir, f.remote), fh)
		_ = fh.Close()
		if err := failed("upload "+f.remote, res, err); err != nil {
			return err
		}
	}
	res, err := in.r.Run(ctx, "cat > "+in.dir+"/SHA256SUMS", strings.NewReader(in.files.sums))
	if err := failed("upload the checksums", res, err); err != nil {
		return err
	}
	if in.files.web == "" {
		return nil
	}
	archive, err := tarWeb(in.files.web)
	if err != nil {
		return err
	}
	res, err = in.r.Run(ctx, fmt.Sprintf("mkdir %s/web && tar -xzf - -C %s/web", in.dir, in.dir), archive)
	return failed("upload the web UI", res, err)
}

func (in *installer) download(ctx context.Context) error {
	if err := in.tempDir(ctx); err != nil {
		return err
	}
	name := "astra-linux-" + in.rep.Arch + ".tar.gz"
	script := fmt.Sprintf(`cd %[1]s
curl -fsSL %[2]s/SHA256SUMS -o SHA256SUMS
curl -fsSL %[2]s/%[3]s -o %[3]s
`, in.dir, strings.TrimSuffix(in.opts.Source.ReleaseURL, "/"), name)
	res, err := in.r.Run(ctx, "sh -s", strings.NewReader("set -eu\n"+script))
	return failed("download the release", res, err)
}

func (in *installer) verify(ctx context.Context) error {
	script := "cd " + in.dir + "\nsha256sum -c SHA256SUMS\n"
	if in.opts.Source.ReleaseURL != "" {
		name := "astra-linux-" + in.rep.Arch + ".tar.gz"
		script = fmt.Sprintf("cd %[1]s\ngrep -E '  %[2]s$' SHA256SUMS | sha256sum -c -\ntar -xzf %[3]s\n", in.dir, strings.ReplaceAll(name, ".", `\.`), name)
	}
	res, err := in.r.Run(ctx, "sh -s", strings.NewReader("set -eu\n"+script))
	if err := failed("verify the checksums", res, err); err != nil {
		return fmt.Errorf("%w; nothing was installed", err)
	}
	return nil
}

// privileged makes a step that runs a script as root with the temporary folder filled in.
func (in *installer) privileged(script string) func(context.Context) error {
	return func(ctx context.Context) error {
		res, err := as(ctx, in.r, in.rep.Root, strings.ReplaceAll(script, "@DIR@", in.dir))
		return failed("run the step", res, err)
	}
}

func (in *installer) service(ctx context.Context) error {
	unit := serviceUnit
	if in.hasWeb() {
		unit = strings.Replace(unit, "@WEB@", "Environment=ASTRA_WEB_DIR=/var/lib/astra/web\n", 1)
	} else {
		unit = strings.Replace(unit, "@WEB@", "", 1)
	}
	script := "cat > /etc/systemd/system/astrad.service <<'ASTRA_UNIT'\n" + unit + "ASTRA_UNIT\nsystemctl daemon-reload\n"
	res, err := as(ctx, in.r, in.rep.Root, script)
	return failed("write the service", res, err)
}

func (in *installer) wait(ctx context.Context) error {
	res, err := in.r.Run(ctx, "sh -s", strings.NewReader(waitScript))
	if err := failed("wait for the daemon", res, err); err != nil {
		return err
	}
	in.version = strings.TrimSpace(string(res.Stdout))
	return nil
}
