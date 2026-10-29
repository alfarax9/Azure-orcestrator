package server

import (
	"bufio"
	"context"
	"fmt"
	"strconv"
	"strings"
)

// Status is how one readiness check came out.
type Status string

// Check statuses. A fail stops the installation; a warn is shown and does not.
const (
	Pass Status = "pass"
	Warn Status = "warn"
	Fail Status = "fail"
)

// Check is one thing the server needs, with what was found and what to do about it.
type Check struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status Status `json:"status"`
	Detail string `json:"detail"`
	Fix    string `json:"fix,omitempty"`
}

// Report is the result of looking at a server (F1-03).
type Report struct {
	Checks []Check `json:"checks"`
	// Arch is the CPU architecture in the names binaries use: amd64 or arm64; empty when unsupported.
	Arch string `json:"arch"`
	// Root says whether the login is root; otherwise steps go through sudo.
	Root bool `json:"root"`
	// Installed is the version of astrad already on the machine, if any.
	Installed string `json:"installed,omitempty"`
	// Running says whether astrad.service is active.
	Running bool `json:"running"`
	// HasDocker says whether Docker is already there.
	HasDocker bool `json:"hasDocker"`
}

// OK reports whether nothing failed, so installing may go ahead.
func (r Report) OK() bool {
	for _, c := range r.Checks {
		if c.Status == Fail {
			return false
		}
	}
	return true
}

// Thresholds from PRD §8.2: 2 vCPU, 4 GB of RAM, 40 GB of disk. A machine reports a little less
// than its nominal RAM, so 3.5 GiB counts as 4 GB.
const (
	okRAMMB, minRAMMB   = 3584, 1792
	okCPUs              = 2
	okDiskGB, minDiskGB = 30, 10
)

// probe prints what Check needs as key=value lines, in one round trip.
const probe = `. /etc/os-release 2>/dev/null || true
echo "uid=$(id -u)"
echo "os_id=${ID:-}"
echo "os_version=${VERSION_ID:-}"
echo "os_name=${PRETTY_NAME:-}"
echo "arch=$(uname -m)"
echo "mem_mb=$(awk '/^MemTotal:/ {printf "%d", $2/1024}' /proc/meminfo 2>/dev/null)"
echo "cpus=$(nproc 2>/dev/null)"
echo "disk_gb=$(df -Pk /var/lib 2>/dev/null | awk 'NR==2 {printf "%d", $4/1048576}')"
if [ -d /run/systemd/system ]; then echo systemd=yes; else echo systemd=no; fi
echo "cgroup=$(stat -fc %T /sys/fs/cgroup 2>/dev/null)"
if [ "$(id -u)" = 0 ] || sudo -n true 2>/dev/null; then echo sudo=yes; else echo sudo=no; fi
for tool in curl git tar sha256sum docker ss; do
  if command -v "$tool" >/dev/null 2>&1; then echo "have_$tool=yes"; else echo "have_$tool=no"; fi
done
echo "ports=$(ss -ltnH 2>/dev/null | awk '{print $4}' | grep -E ':(80|443)$' | tr '\n' ' ')"
echo "installed=$([ -x /usr/local/bin/astrad ] && /usr/local/bin/astrad -version 2>/dev/null || true)"
echo "service=$(systemctl is-active astrad 2>/dev/null || true)"
`

// Probe looks at the machine and judges it.
func Probe(ctx context.Context, r Runner) (Report, error) {
	res, err := r.Run(ctx, "sh -s", strings.NewReader(probe))
	if err := failed("look at the server", res, err); err != nil {
		return Report{}, err
	}
	return Evaluate(parseFacts(string(res.Stdout))), nil
}

func parseFacts(out string) map[string]string {
	facts := map[string]string{}
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		if k, v, ok := strings.Cut(sc.Text(), "="); ok {
			facts[k] = strings.TrimSpace(v)
		}
	}
	return facts
}

// Evaluate judges the facts a probe collected.
func Evaluate(f map[string]string) Report {
	r := Report{Root: f["uid"] == "0", HasDocker: f["have_docker"] == "yes", Running: f["service"] == "active"}
	if v, ok := strings.CutPrefix(f["installed"], "astrad "); ok {
		r.Installed = strings.TrimSpace(v)
	}
	r.add(checkOS(f))
	arch, c := checkArch(f)
	r.Arch = arch
	r.add(c)
	r.add(checkSystemd(f), checkCgroup(f), checkSudo(f))
	r.add(checkRAM(f), checkCPU(f), checkDisk(f), checkPorts(f))
	r.add(checkTool(f, "curl", "curl", Fail, "Checks that the daemon came up, and downloads a release."),
		checkTool(f, "git", "git", Fail, "Clones the projects and makes a worktree for every task."))
	return r
}

func (r *Report) add(cs ...Check) { r.Checks = append(r.Checks, cs...) }

func checkOS(f map[string]string) Check {
	c := Check{ID: "os", Name: "Operating system", Detail: f["os_name"]}
	supported := (f["os_id"] == "ubuntu" && (f["os_version"] == "22.04" || f["os_version"] == "24.04")) ||
		(f["os_id"] == "debian" && f["os_version"] == "12")
	switch {
	case supported:
		c.Status = Pass
	case f["os_id"] == "ubuntu" || f["os_id"] == "debian":
		c.Status, c.Fix = Warn, "Astra is tested on Ubuntu 22.04 and 24.04 and Debian 12. It will probably work here, but nothing promises it."
	default:
		c.Status = Fail
		c.Detail = fmt.Sprintf("%s (not supported)", orDash(f["os_name"]))
		c.Fix = "Use a server running Ubuntu 22.04 or 24.04, or Debian 12."
	}
	return c
}

func checkArch(f map[string]string) (string, Check) {
	c := Check{ID: "arch", Name: "Architecture", Detail: f["arch"]}
	switch f["arch"] {
	case "x86_64", "amd64":
		c.Status = Pass
		return "amd64", c
	case "aarch64", "arm64":
		c.Status = Pass
		return "arm64", c
	}
	c.Status, c.Fix = Fail, "Astra runs on 64-bit Intel or AMD (x86_64) and on 64-bit ARM (aarch64) servers."
	return "", c
}

func checkSystemd(f map[string]string) Check {
	c := Check{ID: "systemd", Name: "systemd", Status: Pass, Detail: "running"}
	if f["systemd"] != "yes" {
		c.Status, c.Detail = Fail, "not running as the init system"
		c.Fix = "Astra runs the daemon and every agent session as systemd units. Use a server image that boots with systemd (not a container without it)."
	}
	return c
}

func checkCgroup(f map[string]string) Check {
	c := Check{ID: "cgroup", Name: "Resource limits (cgroup v2)", Status: Pass, Detail: "available"}
	if f["cgroup"] != "cgroup2fs" {
		c.Status, c.Detail = Fail, orDash(f["cgroup"])
		c.Fix = "Agent sessions get memory and CPU limits through cgroup v2. Ubuntu 22.04 and Debian 12 use it by default; on an older setup, boot with systemd.unified_cgroup_hierarchy=1."
	}
	return c
}

func checkSudo(f map[string]string) Check {
	c := Check{ID: "sudo", Name: "Administrator access", Status: Pass, Detail: "root"}
	switch {
	case f["uid"] == "0":
	case f["sudo"] == "yes":
		c.Detail = "sudo without a password"
	default:
		c.Status, c.Detail = Fail, "this user cannot use sudo without a password"
		c.Fix = "Log in as root, or give the user passwordless sudo: add the line  <user> ALL=(ALL) NOPASSWD:ALL  with visudo."
	}
	return c
}

func checkRAM(f map[string]string) Check {
	mb := atoi(f["mem_mb"])
	c := Check{ID: "ram", Name: "Memory", Detail: fmt.Sprintf("%.1f GB", float64(mb)/1024)}
	switch {
	case mb >= okRAMMB:
		c.Status = Pass
	case mb >= minRAMMB:
		c.Status, c.Fix = Warn, "4 GB is recommended. With less, builds and several agents at once will be slow or run out of memory."
	default:
		c.Status, c.Fix = Fail, "Under 2 GB is too little to run an agent and a build. Use a server with at least 4 GB."
	}
	return c
}

func checkCPU(f map[string]string) Check {
	n := atoi(f["cpus"])
	c := Check{ID: "cpu", Name: "Processors", Detail: fmt.Sprintf("%d", n), Status: Pass}
	if n < okCPUs {
		c.Status, c.Fix = Warn, "2 vCPUs are recommended; with one, an agent and a build compete for it."
	}
	return c
}

func checkDisk(f map[string]string) Check {
	gb := atoi(f["disk_gb"])
	c := Check{ID: "disk", Name: "Free disk", Detail: fmt.Sprintf("%d GB free", gb)}
	switch {
	case gb >= okDiskGB:
		c.Status = Pass
	case gb >= minDiskGB:
		c.Status, c.Fix = Warn, "40 GB is recommended: every worker has its own copy of the code, and builds are cached."
	default:
		c.Status, c.Fix = Fail, "Under 10 GB free is not enough for the code, worktrees and builds. Free some space or use a bigger disk."
	}
	return c
}

func checkPorts(f map[string]string) Check {
	c := Check{ID: "ports", Name: "Ports 80 and 443", Status: Pass, Detail: "free"}
	switch {
	case f["have_ss"] != "yes":
		c.Status, c.Detail = Warn, "could not check (ss is missing)"
		c.Fix = "Install iproute2 so the ports can be checked."
	case strings.TrimSpace(f["ports"]) != "":
		c.Status, c.Detail = Warn, "in use: "+strings.TrimSpace(f["ports"])
		c.Fix = "Previews and production use these ports from a later milestone. Stop the program that holds them (for example nginx or apache) before then."
	}
	return c
}

func checkTool(f map[string]string, key, name string, missing Status, why string) Check {
	c := Check{ID: key, Name: name, Status: Pass, Detail: "installed"}
	if f["have_"+key] != "yes" {
		c.Status, c.Detail = missing, "not installed"
		c.Fix = why + " Install it with: sudo apt-get install -y " + key
	}
	return c
}

func atoi(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}
