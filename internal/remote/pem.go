package remote

import (
	"encoding/pem"
	"fmt"
	"io"
)

func writePEM(w io.Writer, block *pem.Block) error {
	if err := pem.Encode(w, block); err != nil {
		return fmt.Errorf("write the private key: %w", err)
	}
	return nil
}
