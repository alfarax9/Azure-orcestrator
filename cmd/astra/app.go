package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"

	"github.com/alfarax9/astra/internal/remote"
	"github.com/alfarax9/astra/internal/server"
)

// app holds what every command needs: where the config lives, and the terminal.
type app struct {
	config   string
	registry server.Registry
	hosts    remote.KnownHosts
	in       *bufio.Reader
	out, err io.Writer
}

var errUsage = errors.New("wrong arguments")

func isUsage(err error) bool { return errors.Is(err, errUsage) }

func newApp() (*app, error) {
	dir := os.Getenv("ASTRA_CONFIG_DIR")
	if dir == "" {
		base := os.Getenv("XDG_CONFIG_HOME")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("find the home folder: %w", err)
			}
			base = filepath.Join(home, ".config")
		}
		dir = filepath.Join(base, "astra")
	}
	return &app{
		config:   dir,
		registry: server.Registry{Path: filepath.Join(dir, "servers.json")},
		hosts:    remote.KnownHosts{Path: filepath.Join(dir, "known_hosts")},
		in:       bufio.NewReader(os.Stdin),
		out:      os.Stdout,
		err:      os.Stderr,
	}, nil
}

// Writing to the terminal can only fail when there is no terminal, and nothing could be done about it.
func (a *app) errf(format string, args ...any) { _, _ = fmt.Fprintf(a.err, format, args...) }
func (a *app) errln(args ...any)               { _, _ = fmt.Fprintln(a.err, args...) }
func (a *app) outf(format string, args ...any) { _, _ = fmt.Fprintf(a.out, format, args...) }
func (a *app) outln(args ...any)               { _, _ = fmt.Fprintln(a.out, args...) }
func (a *app) outs(s string)                   { _, _ = io.WriteString(a.out, s) }

func (a *app) keyPath(name string) string { return filepath.Join(a.config, "keys", name) }

// confirm asks a yes or no question on the terminal; anything but yes is no.
func (a *app) confirm(question string) bool {
	a.errf("%s [y/N] ", question)
	line, _ := a.in.ReadString('\n')
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes"
}

// passphrase asks for a key's passphrase without echoing it.
func (a *app) passphrase() ([]byte, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return nil, errors.New("the SSH key needs a passphrase, and there is no terminal to ask on")
	}
	a.errf("Passphrase of the SSH key: ")
	pass, err := term.ReadPassword(fd)
	a.errln()
	return pass, err
}

// key makes a new ed25519 key and prints the public key to add to the server (F1-02).
func (a *app) key(args []string) error {
	if len(args) == 0 || args[0] != "new" || len(args) > 2 {
		return errUsage
	}
	name := "astra"
	if len(args) == 2 {
		name = args[1]
	}
	if strings.ContainsAny(name, `/\.`) || name == "" {
		return fmt.Errorf("%w: a key name is letters, digits and hyphens", errUsage)
	}
	pub, err := remote.GenerateKey(a.keyPath(name), "astra")
	if err != nil {
		return err
	}
	a.errf("Key saved in %s. Add this public key to ~/.ssh/authorized_keys of the user on the server:\n\n", a.keyPath(name))
	a.outs(pub)
	return nil
}
