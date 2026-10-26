package domain

import (
	"fmt"
	"strconv"
	"time"
)

// Timestamp is a time that travels as milliseconds since the epoch, the form the web
// client reads. encoding/json/v2 has a format tag for this, but it is still experimental.
type Timestamp struct{ time.Time }

// MarshalJSON encodes the time as a JSON number of milliseconds.
func (t Timestamp) MarshalJSON() ([]byte, error) {
	return strconv.AppendInt(nil, t.UnixMilli(), 10), nil
}

// UnmarshalJSON decodes a JSON number of milliseconds.
func (t *Timestamp) UnmarshalJSON(data []byte) error {
	ms, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return fmt.Errorf("timestamp %s is not milliseconds: %w", data, err)
	}
	t.Time = time.UnixMilli(ms)
	return nil
}
