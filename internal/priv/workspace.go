package priv

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/alfarax9/astra/internal/domain"
)

const (
	groupadd = "/usr/sbin/groupadd"
	useradd  = "/usr/sbin/useradd"
)

// workspaceAdd creates the agent user of a workspace, its folders and the one sudo rule that
// lets the daemon run commands as that user. It can be repeated: what exists is kept.
func (h Helper) workspaceAdd(ctx context.Context, args []string) error {
	if len(args) != 1 || !domain.ValidSlug(args[0]) {
		return fmt.Errorf("%w: workspace-add takes one slug of lowercase letters, digits and hyphens", ErrUsage)
	}
	slug := args[0]
	agent := agentUser(slug)
	if !h.Host.GroupExists(h.Config.DaemonUser) {
		return fmt.Errorf("group %s does not exist: install the daemon first", h.Config.DaemonUser)
	}
	if err := h.ensureAgentUser(ctx, slug, agent); err != nil {
		return err
	}
	// The daemon's group may walk through base to the sessions folder and use the sockets
	// there, and nothing else; home and work stay out of its reach (ADR 0005).
	for _, d := range []Dir{
		{h.Config.StateDir, "root", "root", 0o755},
		{h.Config.StateDir + "/agents", "root", "root", 0o755},
		{h.Config.base(slug), agent, h.Config.DaemonUser, 0o710},
		{h.Config.sessionsDir(slug), agent, h.Config.DaemonUser, 0o770 | fs.ModeSetgid},
		{h.Config.workDir(slug), agent, agent, 0o700},
		{h.Config.homeDir(slug), agent, agent, 0o700},
	} {
		if err := h.Host.Install(d); err != nil {
			return fmt.Errorf("prepare %s: %w", d.Path, err)
		}
	}
	rule := fmt.Sprintf("# Managed by astra-priv. Do not edit.\n%[1]s ALL=(%[2]s) NOPASSWD: ALL\nDefaults:%[1]s !requiretty\n",
		h.Config.DaemonUser, agent)
	if err := h.Host.WriteSudoers(ctx, h.Config.SudoersDir, agent, rule); err != nil {
		return fmt.Errorf("allow the daemon to run as %s: %w", agent, err)
	}
	return nil
}

// ensureAgentUser creates the user and its group, or accepts one that already exists as
// long as it has no groups of its own beyond its primary one: a user in docker, sudo, adm or
// the daemon's group would break the isolation, so that is an error, never a quiet fix.
func (h Helper) ensureAgentUser(ctx context.Context, slug, agent string) error {
	existing, err := h.Host.LookupAgent(agent)
	switch {
	case err == nil:
		if len(existing.ExtraGroups) > 0 {
			return fmt.Errorf("user %s already exists and is in groups %s; remove it from them first",
				agent, strings.Join(existing.ExtraGroups, ", "))
		}
		return nil
	case !errors.Is(err, fs.ErrNotExist):
		return fmt.Errorf("look up user %s: %w", agent, err)
	}
	if !h.Host.GroupExists(agent) {
		if out, err := h.Host.Exec(ctx, groupadd, "--system", agent); err != nil {
			return fmt.Errorf("create group %s: %w: %s", agent, err, strings.TrimSpace(string(out)))
		}
	}
	out, err := h.Host.Exec(ctx, useradd, "--system", "--gid", agent, "--home-dir", h.Config.homeDir(slug),
		"--no-create-home", "--shell", "/bin/bash", agent)
	if err != nil {
		return fmt.Errorf("create user %s: %w: %s", agent, err, strings.TrimSpace(string(out)))
	}
	return nil
}
