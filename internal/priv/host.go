package priv

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const visudo = "/usr/sbin/visudo"

// SystemHost changes the machine it runs on. It needs root for everything except ResolveDir.
type SystemHost struct{}

// Exec runs argv with a fixed, minimal environment.
func (SystemHost) Exec(ctx context.Context, argv ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...) //nolint:gosec // argv is built from validated fields in this package
	cmd.Env = []string{"PATH=/usr/sbin:/usr/bin:/sbin:/bin", "LANG=C"}
	cmd.Dir = "/"
	return cmd.CombinedOutput()
}

// LookupAgent reads the user and the groups it belongs to besides its own.
func (SystemHost) LookupAgent(name string) (Agent, error) {
	u, err := user.Lookup(name)
	if err != nil {
		var unknown user.UnknownUserError
		if errors.As(err, &unknown) {
			return Agent{}, fmt.Errorf("user %s: %w", name, fs.ErrNotExist)
		}
		return Agent{}, err
	}
	ids, err := u.GroupIds()
	if err != nil {
		return Agent{}, fmt.Errorf("list groups of %s: %w", name, err)
	}
	var a Agent
	for _, id := range ids {
		if id == u.Gid {
			continue
		}
		g, err := user.LookupGroupId(id)
		if err != nil {
			a.ExtraGroups = append(a.ExtraGroups, id)
			continue
		}
		a.ExtraGroups = append(a.ExtraGroups, g.Name)
	}
	return a, nil
}

// GroupExists reports whether the group exists.
func (SystemHost) GroupExists(name string) bool {
	_, err := user.LookupGroup(name)
	return err == nil
}

// Install creates the folder and sets its owner and mode. The folder is opened without
// following symbolic links and changed through that open handle: an agent owns the
// folders below its workspace and could otherwise swap one for a link and make root change
// something else.
func (SystemHost) Install(d Dir) error {
	uid, gid, err := ids(d.Owner, d.Group)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(d.Path), 0o755); err != nil { //nolint:gosec // parents of the state folder are public
		return err
	}
	if err := os.Mkdir(d.Path, 0o700); err != nil && !errors.Is(err, fs.ErrExist) {
		return err
	}
	f, err := os.OpenFile(d.Path, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_DIRECTORY, 0)
	if err != nil {
		return fmt.Errorf("open %s without following links: %w", d.Path, err)
	}
	defer func() { _ = f.Close() }()
	if err := f.Chown(uid, gid); err != nil {
		return err
	}
	return f.Chmod(d.Mode)
}

func ids(owner, group string) (uid, gid int, err error) {
	u, err := user.Lookup(owner)
	if err != nil {
		return 0, 0, fmt.Errorf("look up user %s: %w", owner, err)
	}
	g, err := user.LookupGroup(group)
	if err != nil {
		return 0, 0, fmt.Errorf("look up group %s: %w", group, err)
	}
	if uid, err = strconv.Atoi(u.Uid); err != nil {
		return 0, 0, err
	}
	gid, err = strconv.Atoi(g.Gid)
	return uid, gid, err
}

// WriteSudoers writes the rule to a file sudo ignores (its name starts with a dot), has
// visudo check it, and renames it into place, so sudo never sees a broken file.
func (h SystemHost) WriteSudoers(ctx context.Context, dir, name, content string) error {
	tmp := filepath.Join(dir, "."+name+".tmp")
	_ = os.Remove(tmp)
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o440) //nolint:gosec // sudoers files are 0440
	if err != nil {
		return err
	}
	_, werr := f.WriteString(content)
	if cerr := f.Close(); werr == nil {
		werr = cerr
	}
	if werr != nil {
		_ = os.Remove(tmp)
		return werr
	}
	if out, err := h.Exec(ctx, visudo, "-cf", tmp); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("visudo rejected the rule: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return os.Rename(tmp, filepath.Join(dir, name))
}

// ResolveDir returns the real path of root/rel after following links, and refuses anything
// that is not a folder inside root.
func (SystemHost) ResolveDir(root, rel string) (string, error) {
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	real, err := filepath.EvalSymlinks(filepath.Join(realRoot, rel))
	if err != nil {
		return "", err
	}
	if real != realRoot && !strings.HasPrefix(real, realRoot+string(filepath.Separator)) {
		return "", fmt.Errorf("%s leads outside the work folder", rel)
	}
	info, err := os.Stat(real)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a folder", rel)
	}
	return real, nil
}
