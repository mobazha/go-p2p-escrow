package ports

// UTXOWallet provides the low-level UTXO operations needed by the
// UTXO escrow adapter. This is a subset of what a full wallet does —
// only the escrow-relevant methods.
type UTXOWallet interface {
	// CreateMultisigAddress generates an M-of-N P2WSH address from the
	// given public keys, chaincode, and threshold.
	// Returns the address string and the redeem script.
	CreateMultisigAddress(keys [][]byte, chaincode []byte, threshold int) (address string, redeemScript []byte, err error)

	// CreateMultisigWithTimeout generates a P2WSH address that can be
	// spent by M-of-N normally, or by a single timeout key after the
	// specified number of blocks.
	// Returns the address string and the redeem script.
	CreateMultisigWithTimeout(keys [][]byte, chaincode []byte, threshold int, timeoutBlocks uint32, timeoutKey []byte) (address string, redeemScript []byte, err error)

	// SignMultisigTransaction signs a multisig spending transaction
	// with the given private key and redeem script.
	// Returns one signature per input.
	SignMultisigTransaction(rawTx []byte, privateKey []byte, redeemScript []byte) ([]InputSignature, error)

	// BuildAndBroadcast assembles a fully-signed multisig transaction
	// from collected signatures and broadcasts it.
	// Returns the transaction hash.
	BuildAndBroadcast(rawTx []byte, signatures [][]InputSignature, redeemScript []byte) (txHash string, err error)

	// EstimateEscrowFee estimates the fee for an escrow spend transaction.
	EstimateEscrowFee(threshold int, numOutputs int, feeLevel int) (satoshis int64, err error)
}

// InputSignature is a signature for a single transaction input.
type InputSignature struct {
	Index     int
	Signature []byte
}
