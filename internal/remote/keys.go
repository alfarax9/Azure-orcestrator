package remote

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

// GenerateKey makes a new ed25519 key at path (mode 0600) and returns the public key in the
// form authorized_keys takes, for the person to add to the server (F1-02). It never replaces a
// key that already exists.
func GenerateKey(path, comment string) (string, error) {
	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("%s already exists; pick another name or use that key", path)
	}
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", fmt.Errorf("generate a key: %w", err)
	}
	block, err := ssh.MarshalPrivateKey(priv, comment)
	if err != nil {
		return "", fmt.Errorf("encode the private key: %w", err)
	}
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return "", fmt.Errorf("encode the public key: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", fmt.Errorf("create the key folder: %w", err)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600) //nolint:gosec // the new key file, mode 0600
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return "", fmt.Errorf("%s already exists", path)
		}
		return "", fmt.Errorf("create the key file: %w", err)
	}
	defer func() { _ = f.Close() }()
	if err := writePEM(f, block); err != nil {
		return "", err
	}
	line := string(ssh.MarshalAuthorizedKey(sshPub))
	if comment != "" {
		line = line[:len(line)-1] + " " + comment + "\n"
	}
	return line, os.WriteFile(path+".pub", []byte(line), 0o644) //nolint:gosec // a public key
}
