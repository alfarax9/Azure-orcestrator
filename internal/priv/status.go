package priv

import (
	"encoding/json/v2"
	"fmt"
	"io"
	"path/filepath"

	"github.com/alfarax9/astra/internal/domain"
)

// agentStatus prints which agents the workspace's agent user is signed in to, as JSON. The
// daemon may not open the agent's home (F3-03), so the answer comes from here, and it is only
// whether the login files exist.
func (h Helper) agentStatus(args []string, out io.Writer) error {
	if len(args) != 1 || !domain.ValidSlug(args[0]) {
		return fmt.Errorf("%w: agent-status takes one workspace slug", ErrUsage)
	}
	status := map[string]bool{}
	for agent, file := range credentialFiles {
		status[agent] = h.Host.FileExists(filepath.Join(h.Config.homeDir(args[0]), file))
	}
	b, err := json.Marshal(status, json.Deterministic(true))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, string(b))
	return err
}
