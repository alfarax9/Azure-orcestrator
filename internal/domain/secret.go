package domain

import "log/slog"

const redacted = "[redacted]"

// Secret holds a value that must never reach a log, an error message, JSON or an agent:
// Cloudflare tokens, production environment variables, pairing tokens (N-11). Printing it
// in any way shows [redacted]; the real value comes out only through Reveal, which makes
// every use easy to find.
type Secret struct{ value string }

// NewSecret wraps v.
func NewSecret(v string) Secret { return Secret{value: v} }

// Reveal returns the real value. Call it only where the value is actually used.
func (s Secret) Reveal() string { return s.value }

// IsSet reports whether the secret holds a value.
func (s Secret) IsSet() bool { return s.value != "" }

// String implements fmt.Stringer.
func (s Secret) String() string { return redacted }

// GoString keeps %#v from printing the value.
func (s Secret) GoString() string { return redacted }

// LogValue implements slog.LogValuer.
func (s Secret) LogValue() slog.Value { return slog.StringValue(redacted) }

// MarshalText keeps the value out of JSON and any other text encoding.
func (s Secret) MarshalText() ([]byte, error) { return []byte(redacted), nil }

// Last4 returns the last four characters, for showing that a secret is saved without
// showing it (DESIGN_SYSTEM.md section 8.2). Shorter secrets show nothing.
func (s Secret) Last4() string {
	const n = 4
	if len(s.value) <= n*2 {
		return ""
	}
	return s.value[len(s.value)-n:]
}
