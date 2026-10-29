// Package auth pairs clients with the daemon and recognises them afterwards (F1-05,
// ADR 0006). A pairing token is made on the server, handed to the client over SSH, and works
// once; the client trades it for a device token. Only hashes of tokens are stored, and tokens
// are domain.Secret while in memory, so none reaches a log.
package auth
