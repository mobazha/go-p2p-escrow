// Package escrow provides a privacy-preserving, multi-chain P2P escrow SDK.
//
// It handles the full lifecycle of escrowed payments: setup, funding,
// release, refund, and dispute resolution across UTXO chains, Monero,
// and (future) EVM/Solana. Every chain adapter implements the [Escrow]
// interface; the [Registry] dispatches by [ChainType].
package escrow

// ChainType identifies a blockchain network.
type ChainType string

const (
	ChainBitcoin     ChainType = "BTC"
	ChainBitcoinCash ChainType = "BCH"
	ChainLitecoin    ChainType = "LTC"
	ChainZCash       ChainType = "ZEC"
	ChainEthereum    ChainType = "ETH"
	ChainSolana      ChainType = "SOL"
	ChainMonero      ChainType = "XMR"
)

// IsUTXO reports whether the chain uses the UTXO transaction model.
func (c ChainType) IsUTXO() bool {
	switch c {
	case ChainBitcoin, ChainBitcoinCash, ChainLitecoin, ChainZCash:
		return true
	}
	return false
}

// String returns the chain's ticker symbol.
func (c ChainType) String() string { return string(c) }
