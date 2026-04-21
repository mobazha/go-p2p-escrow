package ports

import "context"

// ChainClient provides read-only access to on-chain data needed
// for escrow verification and fee estimation.
type ChainClient interface {
	// GetBalance returns the confirmed balance in the chain's smallest
	// unit for the given address.
	GetBalance(ctx context.Context, address string) (confirmed int64, unconfirmed int64, err error)

	// GetTransaction returns the raw transaction data for the given hash.
	GetTransaction(ctx context.Context, txHash string) (*TxInfo, error)
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
