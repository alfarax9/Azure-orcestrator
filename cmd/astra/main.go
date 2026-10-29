// Command astra is the Astra command line: it sets up SSH keys, adds a server, checks that it
// can run Astra, installs the daemon on it, and pairs this machine with it (ADR 0007). The same
// installation can run on the server itself with `astra install`.
package main

import (
	"fmt"
	"os"
)

// version is set at release time with -ldflags "-X main.version=1.0.0".
var version = "dev"

const usage = `usage: astra <command>
  key new [name]                              make an SSH key and print its public key
  server add <name> --host H [--user U] [--port N] [--key PATH]
  server list | remove <name>
  server check <name>                         can this server run Astra?
  server install <name> (--binary-dir DIR | --release-url URL) [--docker] [--yes]
  server pair <name>                          pair this machine with the daemon
  server tunnel <name> [--port N]             reach the daemon at http://127.0.0.1:N
  install (--binary-dir DIR | --release-url URL) [--docker] [--yes]
                                              install the daemon on this machine (run it on the server)
  -version`

func main() { os.Exit(run(os.Args[1:])) }

func run(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}
	app, err := newApp()
	if err != nil {
		fmt.Fprintln(os.Stderr, "astra:", err)
		return 1
	}
	var cmdErr error
	switch args[0] {
	case "-version", "--version":
		fmt.Println("astra", version)
		return 0
	case "key":
		cmdErr = app.key(args[1:])
	case "server":
		cmdErr = app.serverCommand(args[1:])
	case "install":
		cmdErr = app.installHere(args[1:])
	default:
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}
	return finish(cmdErr)
}

// finish turns a command's error into the exit status: a usage mistake is 2, any other failure 1.
func finish(err error) int {
	switch {
	case err == nil:
		return 0
	case isUsage(err):
		fmt.Fprintln(os.Stderr, "astra:", err)
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}
	fmt.Fprintln(os.Stderr, "astra:", err)
	return 1
}
