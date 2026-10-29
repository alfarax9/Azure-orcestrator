// Command astrad is the Astra daemon. It runs on the user's VPS, stores facts about
// workers, previews and deployments, and drives agents, previews, builds and deploys.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/alfarax9/astra/internal/pty"
)

const mainUsage = `usage: astrad <command>
  serve                the daemon: the API on 127.0.0.1:7437 ($ASTRA_LISTEN), the web UI in $ASTRA_WEB_DIR, and a watch on every running agent
  pair                 print a pairing token for a new device (valid 10 minutes, works once)
  devices [revoke ID]  list paired devices, or cut one off
  workspace|project|worker ...   manage objects from the command line (astrad worker for details)
  session|events ...   the early spike commands for sessions and transcripts
  -version`

// version is set at release time with -ldflags "-X main.version=1.0.0".
var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case pty.HostCommand:
			os.Exit(pty.HostMainWith(os.Args[2:], os.Environ(), followers()))
		case "session":
			os.Exit(runSession(os.Args[2:]))
		case "events":
			os.Exit(runEvents(os.Args[2:]))
		case "serve":
			os.Exit(runServe(os.Args[2:]))
		case "pair":
			os.Exit(runPair())
		case "devices":
			os.Exit(runDevices(os.Args[2:]))
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
	fmt.Fprintln(os.Stderr, mainUsage)
	os.Exit(2)
}
