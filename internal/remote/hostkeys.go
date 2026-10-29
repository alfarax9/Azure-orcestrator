package remote

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// UnknownHostError means this is the first connection to the host. Fingerprint is what a
// person compares with what the server's provider shows before trusting it.
type UnknownHostError struct {
	Host        string
	KeyType     string
	Fingerprint string
}

func (e *UnknownHostError) Error() string {
	return fmt.Sprintf("the host key of %s is not known yet (%s %s)", e.Host, e.KeyType, e.Fingerprint)
}

// ChangedHostKeyError means the host answers with a different key than the one remembered:
// either the server was reinstalled, or someone is in the way. It is never overridden here.
type ChangedHostKeyError struct {
	Host        string
	Fingerprint string
	Known       string
}

func (e *ChangedHostKeyError) Error() string {
	return fmt.Sprintf("the host key of %s changed: it is now %s, and %s was remembered. "+
		"If you reinstalled the server, remove its line from the known hosts file and try again; otherwise do not continue",
		e.Host, e.Fingerprint, e.Known)
}

// KnownHosts is the file of host keys Astra has been told to trust, in OpenSSH's format.
type KnownHosts struct{ Path string }

// callback checks a host key against the file.
func (k KnownHosts) callback() (ssh.HostKeyCallback, error) {
	if err := os.MkdirAll(filepath.Dir(k.Path), 0o700); err != nil {
		return nil, fmt.Errorf("create known hosts folder: %w", err)
	}
	f, err := os.OpenFile(k.Path, os.O_CREATE|os.O_RDONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open known hosts: %w", err)
	}
	_ = f.Close()
	check, err := knownhosts.New(k.Path)
	if err != nil {
		return nil, fmt.Errorf("read known hosts: %w", err)
	}
	return func(host string, remote net.Addr, key ssh.PublicKey) error {
		err := check(host, remote, key)
		var ke *knownhosts.KeyError
		switch {
		case err == nil:
			return nil
		case errors.As(err, &ke) && len(ke.Want) == 0:
			return &UnknownHostError{Host: host, KeyType: key.Type(), Fingerprint: ssh.FingerprintSHA256(key)}
		case errors.As(err, &ke):
			return &ChangedHostKeyError{Host: host, Fingerprint: ssh.FingerprintSHA256(key), Known: ssh.FingerprintSHA256(ke.Want[0].Key)}
		}
		return err
	}, nil
}

// Trust remembers the key of host, after a person confirmed its fingerprint.
func (k KnownHosts) Trust(host string, key ssh.PublicKey) error {
	f, err := os.OpenFile(k.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open known hosts: %w", err)
	}
	line := knownhosts.Line([]string{host}, key)
	if _, err := f.WriteString(line + "\n"); err != nil {
		_ = f.Close()
		return fmt.Errorf("remember host key: %w", err)
	}
	return f.Close()
}

// hostLabel is how the known hosts file names a host and port.
func hostLabel(host string, port int) string {
	return knownhosts.Normalize(net.JoinHostPort(strings.TrimSpace(host), fmt.Sprint(port)))
}
