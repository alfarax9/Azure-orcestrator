package server

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/alfarax9/astra/internal/domain"
)

// Exchange trades a pairing token for a device token, by calling the daemon's pairing endpoint
// through whatever reaches it: an SSH tunnel from a laptop, or the loopback on the server itself.
type Exchange func(ctx context.Context, pairingToken, deviceName string) (deviceToken string, err error)

// tokenPattern is what a pairing token looks like (URL-safe base64), so that anything else
// the command printed is not mistaken for one.
var tokenPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{20,128}$`)

// The command is fixed text, and names the database so it does not depend on whose HOME the
// login has. It runs as the daemon's user, the only one that may open the database.
const (
	pairAsRoot   = "runuser -u astrad -- env ASTRA_DB=/var/lib/astrad/astra.db /usr/local/bin/astrad pair"
	pairAsSudoer = "sudo -n -u astrad env ASTRA_DB=/var/lib/astrad/astra.db /usr/local/bin/astrad pair"
)

// PairingToken makes a pairing token on the server (it works once and expires after ten
// minutes). Whoever calls it decides where the token goes; it is a domain.Secret so that it
// cannot reach a log by accident.
func PairingToken(ctx context.Context, r Runner, root bool) (domain.Secret, error) {
	cmd := pairAsSudoer
	if root {
		cmd = pairAsRoot
	}
	res, err := r.Run(ctx, cmd, nil)
	if err := failed("make a pairing token", res, err); err != nil {
		return domain.Secret{}, fmt.Errorf("%w (is the daemon installed? run the install first)", err)
	}
	token := domain.NewSecret(strings.TrimSpace(string(res.Stdout)))
	if !tokenPattern.MatchString(token.Reveal()) {
		return domain.Secret{}, errors.New("the server did not print a pairing token")
	}
	return token, nil
}

// Pair makes a pairing token on the server and trades it for a device token (F1-05). The pairing
// token exists only in this function's variables: it is read from the command's output, passed to
// exchange, and never printed or kept.
func Pair(ctx context.Context, r Runner, root bool, deviceName string, exchange Exchange) (domain.Secret, error) {
	pairing, err := PairingToken(ctx, r, root)
	if err != nil {
		return domain.Secret{}, err
	}
	device, err := exchange(ctx, pairing.Reveal(), deviceName)
	if err != nil {
		return domain.Secret{}, fmt.Errorf("trade the pairing token for a device token: %w", err)
	}
	if device == "" {
		return domain.Secret{}, errors.New("the daemon did not give a device token")
	}
	return domain.NewSecret(device), nil
}
