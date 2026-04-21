package utxo

import (
	"crypto/sha256"
	"fmt"

	btcec "github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

// CreateMultisigScript builds an M-of-N OP_CHECKMULTISIG redeem script.
// The public keys must be in a canonical, deterministic order (sorted by
// compressed serialization) so that all parties produce the same script.
func CreateMultisigScript(keys []*btcec.PublicKey, threshold int) ([]byte, error) {
	if len(keys) < threshold {
		return nil, fmt.Errorf("need at least %d keys, got %d", threshold, len(keys))
	}
	if len(keys) > 8 {
		return nil, fmt.Errorf("max 8 public keys, got %d", len(keys))
	}

	builder := txscript.NewScriptBuilder()
	builder.AddInt64(int64(threshold))
	for _, key := range keys {
		builder.AddData(key.SerializeCompressed())
	}
	builder.AddInt64(int64(len(keys)))
	builder.AddOp(txscript.OP_CHECKMULTISIG)

	return builder.Script()
}

// ScriptToAddress converts a redeem script to a chain address.
// P2WSH: SHA256(script) → bech32 witness address (BTC, LTC).
// P2SH:  HASH160(script) → base58 script address (BCH, ZEC).
func ScriptToAddress(redeemScript []byte, params *chaincfg.Params, addrType AddressType) (string, error) {
	switch addrType {
	case P2WSH:
		witnessProgram := sha256.Sum256(redeemScript)
		addr, err := btcutil.NewAddressWitnessScriptHash(witnessProgram[:], params)
		if err != nil {
			return "", err
		}
		return addr.String(), nil
	case P2SH:
		addr, err := btcutil.NewAddressScriptHash(redeemScript, params)
		if err != nil {
			return "", err
		}
		return addr.String(), nil
	default:
		return "", fmt.Errorf("unsupported address type: %d", addrType)
	}
}

// GetPayToAddrScript returns the scriptPubKey for a given address string.
func GetPayToAddrScript(addr string, params *chaincfg.Params) ([]byte, error) {
	address, err := btcutil.DecodeAddress(addr, params)
	if err != nil {
		return nil, fmt.Errorf("decode address %q: %w", addr, err)
	}
	return txscript.PayToAddrScript(address)
}
