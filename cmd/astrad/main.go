// Command astrad is the Astra daemon. It runs on the user's VPS, stores facts about
// workers, previews and deployments, and drives agents, previews, builds and deploys.
package main

import (
	"flag"
	"fmt"
	"os"
)

// version is set at release time with -ldflags "-X main.version=1.0.0".
var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println("astrad", version)
		return
	}
	fmt.Fprintln(os.Stderr, "astrad: the daemon is not built yet (M0 spikes come first, see docs/PRD.md)")
	os.Exit(1)
}
