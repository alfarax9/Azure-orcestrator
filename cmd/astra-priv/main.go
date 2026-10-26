// Command astra-priv is the one program Astra runs as root. The daemon calls it through
// sudo for the three things that need root: workspace-add, session-start and session-stop
// (docs/adr/0005). Everything it accepts is validated in internal/priv.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/alfarax9/astra/internal/priv"
)

func main() { os.Exit(run()) }

func run() int {
	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "astra-priv: must run as root (the daemon calls it through sudo)")
		return 1
	}
	syscall.Umask(0o022)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	h := priv.Helper{Config: priv.DefaultConfig(), Host: priv.SystemHost{}}
	if err := h.Run(ctx, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "astra-priv:", err)
		if errors.Is(err, priv.ErrUsage) {
			return 2
		}
		return 1
	}
	return 0
}
