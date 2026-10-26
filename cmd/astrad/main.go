// Command astrad is the Astra daemon. It runs on the user's VPS, stores facts about
// workers, previews and deployments, and drives agents, previews, builds and deploys.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/alfarax9/astra/internal/pty"
)

// version is set at release time with -ldflags "-X main.version=1.0.0".
var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case pty.HostCommand:
			os.Exit(pty.HostMain(os.Args[2:], os.Environ()))
		case "session":
			os.Exit(runSession(os.Args[2:]))
		case "events":
			os.Exit(runEvents(os.Args[2:]))
		case "workspace", "project", "worker":
			os.Exit(runObjects(os.Args[1], os.Args[2:]))
		}
	}
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println("astrad", version)
		return
	}
	fmt.Fprintln(os.Stderr, "astrad: the daemon is not built yet (M0 spikes come first, see docs/PRD.md)")
	os.Exit(1)
}
