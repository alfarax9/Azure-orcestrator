package server

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
)

// Entry is a server this laptop knows (the Server of PRD §7, as far as M1 needs).
type Entry struct {
	Name    string `json:"name"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
	User    string `json:"user"`
	KeyPath string `json:"keyPath"`
	Role    string `json:"role"` // "all-in-one" until agent and production servers exist
	// DeviceToken is what the daemon gave this laptop when it paired. It is the one secret kept
	// here, in a file only the owner can read; the desktop app keeps it in the OS keychain instead.
	DeviceToken string `json:"deviceToken,omitempty"`
}

// Paired reports whether this laptop has a device token for the server.
func (e Entry) Paired() bool { return e.DeviceToken != "" }

var (
	namePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,30}$`)
	hostPattern = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9.:-]{0,251}[A-Za-z0-9])?$`)
	userPattern = regexp.MustCompile(`^[a-z_][a-z0-9_-]{0,31}$`)
)

// ErrInvalid means a name, host or user was refused.
var ErrInvalid = errors.New("invalid")

// Validate checks what goes into commands and file names.
func (e Entry) Validate() error {
	switch {
	case !namePattern.MatchString(e.Name):
		return fmt.Errorf("%w: a server name is lowercase letters, digits and hyphens, starting with a letter", ErrInvalid)
	case !hostPattern.MatchString(e.Host):
		return fmt.Errorf("%w: %q is not a host name or address", ErrInvalid, e.Host)
	case !userPattern.MatchString(e.User):
		return fmt.Errorf("%w: %q is not a user name", ErrInvalid, e.User)
	case e.Port < 1 || e.Port > 65535:
		return fmt.Errorf("%w: the port is 1 to 65535", ErrInvalid)
	case e.KeyPath == "":
		return fmt.Errorf("%w: choose an SSH key", ErrInvalid)
	}
	return nil
}

// Registry is the file of servers, in the client's config folder.
type Registry struct{ Path string }

// ErrNotFound means no server has that name.
var ErrNotFound = errors.New("no such server")

// List returns the servers in the order they were added.
func (g Registry) List() ([]Entry, error) {
	data, err := os.ReadFile(g.Path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read servers: %w", err)
	}
	var list []Entry
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("read %s: %w", g.Path, err)
	}
	return list, nil
}

// Get finds a server by name.
func (g Registry) Get(name string) (Entry, error) {
	list, err := g.List()
	if err != nil {
		return Entry{}, err
	}
	for _, e := range list {
		if e.Name == name {
			return e, nil
		}
	}
	return Entry{}, fmt.Errorf("%w: %s", ErrNotFound, name)
}

// Save adds the server, or replaces the one with the same name.
func (g Registry) Save(e Entry) error {
	if err := e.Validate(); err != nil {
		return err
	}
	list, err := g.List()
	if err != nil {
		return err
	}
	if i := slices.IndexFunc(list, func(x Entry) bool { return x.Name == e.Name }); i >= 0 {
		list[i] = e
	} else {
		list = append(list, e)
	}
	return g.write(list)
}

// Remove forgets a server on this laptop. Nothing on the server changes.
func (g Registry) Remove(name string) error {
	list, err := g.List()
	if err != nil {
		return err
	}
	i := slices.IndexFunc(list, func(x Entry) bool { return x.Name == name })
	if i < 0 {
		return fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return g.write(slices.Delete(list, i, i+1))
}

// write replaces the file in one step, so an interrupted save never leaves half a file, and it is
// never readable by anyone else.
func (g Registry) write(list []Entry) error {
	if list == nil {
		list = []Entry{}
	}
	data, err := json.Marshal(list)
	if err != nil {
		return fmt.Errorf("write servers: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(g.Path), 0o700); err != nil {
		return fmt.Errorf("create the config folder: %w", err)
	}
	tmp := g.Path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write servers: %w", err)
	}
	if err := os.Rename(tmp, g.Path); err != nil {
		return fmt.Errorf("write servers: %w", err)
	}
	return nil
}
