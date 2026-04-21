package utxo

import (
	"errors"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/blockchain"
	btcec "github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/txscript"
)

// CreateTimelockScript builds a redeem script that has two spending paths:
//
//	IF
//	    <threshold> <pubkeys...> <n> OP_CHECKMULTISIG
//	ELSE
//	    <sequence> OP_CHECKSEQUENCEVERIFY OP_DROP <timeoutKey> OP_CHECKSIG
//	ENDIF
//
// The timeout is encoded as a BIP 68 relative sequence lock.
func CreateTimelockScript(keys []*btcec.PublicKey, threshold int, timeout time.Duration, timeoutKey *btcec.PublicKey) ([]byte, error) {
	if len(keys) < threshold {
		return nil, fmt.Errorf("need at least %d keys, got %d", threshold, len(keys))
	}
	if len(keys) > 8 {
		return nil, fmt.Errorf("max 8 public keys, got %d", len(keys))
	}

	// Convert duration to BIP 68 sequence lock (6 blocks per hour approximation).
	sequenceLock := blockchain.LockTimeToSequence(false, uint32(timeout.Hours()*6))

	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_IF)
	builder.AddInt64(int64(threshold))
	for _, key := range keys {
		builder.AddData(key.SerializeCompressed())
	}
	builder.AddInt64(int64(len(keys)))
	builder.AddOp(txscript.OP_CHECKMULTISIG)
	builder.AddOp(txscript.OP_ELSE)
	builder.AddInt64(int64(sequenceLock))
	builder.AddOp(txscript.OP_CHECKSEQUENCEVERIFY)
	builder.AddOp(txscript.OP_DROP)
	builder.AddData(timeoutKey.SerializeCompressed())
	builder.AddOp(txscript.OP_CHECKSIG)
	builder.AddOp(txscript.OP_ENDIF)

	return builder.Script()
}

// IsTimelockScript reports whether the redeem script starts with OP_IF,
// indicating it has a timelock branch.
func IsTimelockScript(redeemScript []byte) bool {
	return len(redeemScript) > 0 && redeemScript[0] == txscript.OP_IF
}

// ParseTimelockSequence extracts the BIP 68 sequence value from a timelock
// redeem script. Returns an error if the script is not a valid timelock.
func ParseTimelockSequence(redeemScript []byte) (uint32, error) {
	if len(redeemScript) < 113 {
		return 0, errors.New("redeem script too short for timelock")
	}
	// After the IF branch: threshold(1) + nPubs*34 + nTotal(1) + OP_CHECKMULTISIG(1) + OP_ELSE(1)
	// For 2-of-3: 1 + 3*34 + 1 + 1 + 1 = 106 → OP_ELSE at index 106
	if redeemScript[106] != txscript.OP_ELSE {
		return 0, errors.New("OP_ELSE not at expected position")
	}

	opcode := redeemScript[107]
	if opcode == 0 {
		return 0, nil
	}
	if opcode >= 81 && opcode <= 96 {
		return uint32(opcode-81) + 1, nil
	}

	if opcode < 1 || opcode > 75 {
		return 0, errors.New("too many bytes pushed for sequence")
	}
	var result int64
	for i := 0; i < int(opcode); i++ {
		result |= int64(redeemScript[108+i]) << uint8(8*i)
	}
	return uint32(result), nil
}
