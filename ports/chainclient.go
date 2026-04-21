package ports

import "context"

// ChainClient provides access to on-chain data needed for escrow
// verification, fee estimation, and transaction broadcast.
type ChainClient interface {
	// GetBalance returns the confirmed balance in the chain's smallest
	// unit for the given address.
	GetBalance(ctx context.Context, address string) (confirmed int64, unconfirmed int64, err error)

	// GetTransaction returns the raw transaction data for the given hash.
	GetTransaction(ctx context.Context, txHash string) (*TxInfo, error)

	// GetFeeRate returns the estimated fee rate in satoshis per byte
	// for a transaction to confirm within targetBlocks.
	GetFeeRate(ctx context.Context, targetBlocks int) (satPerByte int64, err error)

	// Broadcast submits a fully-signed raw transaction to the network
	// and returns the transaction hash.
	Broadcast(ctx context.Context, rawTx []byte) (txHash string, err error)
}

// TxInfo is a minimal transaction representation for escrow verification.
type TxInfo struct {
	Hash          string
	Confirmations int
	BlockHeight   uint64
	Outputs       []TxOutput
}

// TxOutput represents a single output in a transaction.
type TxOutput struct {
	Address string
	Amount  int64
	Index   uint32
}
