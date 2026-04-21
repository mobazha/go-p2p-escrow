// Package ports defines the interfaces that consumers must implement
// to connect go-p2p-escrow to their infrastructure.
package ports

// KeyProvider abstracts access to the master private keys used for
// escrow key derivation. Each chain family has its own master key.
//
// Implementations might read from a file, a KMS/HSM, or a key vault.
// The returned key material must be treated as secret and cleared
// after use.
type KeyProvider interface {
	// EscrowMasterKey returns the BIP32-compatible master private key
	// used to derive per-escrow keys for UTXO chains.
	// Returns the raw 32-byte secp256k1 private key.
	EscrowMasterKey() ([]byte, error)
}
