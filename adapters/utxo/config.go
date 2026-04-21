// Package utxo provides an escrow adapter for UTXO-based blockchains
// (Bitcoin, Litecoin, Bitcoin Cash, Zcash) using P2WSH/P2SH multisig scripts.
package utxo

import (
	"github.com/btcsuite/btcd/chaincfg"

	escrow "github.com/mobazha/go-p2p-escrow"
)

// AddressType determines how a multisig redeem script is wrapped.
type AddressType int

const (
	// P2WSH wraps the script in a SegWit witness script hash (BTC, LTC).
	P2WSH AddressType = iota
	// P2SH wraps the script in a legacy script hash (BCH, ZEC).
	P2SH
)

// ChainConfig holds the chain-specific parameters needed to create
// multisig addresses and sign transactions.
type ChainConfig struct {
	Chain       escrow.ChainType
	Params      *chaincfg.Params
	AddrType    AddressType
	TxVersion   int32 // wire.MsgTx version for normal transactions
	TimelockVer int32 // wire.MsgTx version for timelock (CSV) transactions (typically 2)
}

// Mainnet configurations for each supported UTXO chain.
var (
	BitcoinMainnet = ChainConfig{
		Chain:       escrow.ChainBitcoin,
		Params:      &chaincfg.MainNetParams,
		AddrType:    P2WSH,
		TxVersion:   1,
		TimelockVer: 2,
	}

	BitcoinTestnet = ChainConfig{
		Chain:       escrow.ChainBitcoin,
		Params:      &chaincfg.TestNet3Params,
		AddrType:    P2WSH,
		TxVersion:   1,
		TimelockVer: 2,
	}

	BitcoinRegtest = ChainConfig{
		Chain:       escrow.ChainBitcoin,
		Params:      &chaincfg.RegressionNetParams,
		AddrType:    P2WSH,
		TxVersion:   1,
		TimelockVer: 2,
	}

	LitecoinMainnet = ChainConfig{
		Chain:    escrow.ChainLitecoin,
		Params:   litecoinMainNetParams(),
		AddrType: P2WSH,
		TxVersion:   1,
		TimelockVer: 2,
	}

	ZCashMainnet = ChainConfig{
		Chain:    escrow.ChainZCash,
		Params:   zcashMainNetParams(),
		AddrType: P2SH,
		TxVersion:   1,
		TimelockVer: 2,
	}

	BitcoinCashMainnet = ChainConfig{
		Chain:    escrow.ChainBitcoinCash,
		Params:   bitcoinCashMainNetParams(),
		AddrType: P2SH,
		TxVersion:   1,
		TimelockVer: 2,
	}
)

// litecoinMainNetParams returns chaincfg.Params configured for Litecoin mainnet.
// Only the fields relevant to address encoding are set.
func litecoinMainNetParams() *chaincfg.Params {
	p := chaincfg.Params{
		Name:             "litecoin-mainnet",
		Net:              0xdbb6c0fb,
		PubKeyHashAddrID: 0x30, // L prefix
		ScriptHashAddrID: 0x32, // M prefix
		Bech32HRPSegwit:  "ltc",
		HDPublicKeyID:    [4]byte{0x04, 0x88, 0xb2, 0x1e},
		HDPrivateKeyID:   [4]byte{0x04, 0x88, 0xad, 0xe4},
	}
	return &p
}

// zcashMainNetParams returns chaincfg.Params configured for Zcash mainnet.
func zcashMainNetParams() *chaincfg.Params {
	p := chaincfg.Params{
		Name:             "zcash-mainnet",
		Net:              0x6427e924,
		PubKeyHashAddrID: 0x1C, // t1 prefix (two-byte: 0x1CB8, using first byte)
		ScriptHashAddrID: 0x1C, // t3 prefix (two-byte: 0x1CBD)
		HDPublicKeyID:    [4]byte{0x04, 0x88, 0xb2, 0x1e},
		HDPrivateKeyID:   [4]byte{0x04, 0x88, 0xad, 0xe4},
	}
	return &p
}

// bitcoinCashMainNetParams returns chaincfg.Params for Bitcoin Cash mainnet.
func bitcoinCashMainNetParams() *chaincfg.Params {
	p := chaincfg.MainNetParams
	p.Name = "bitcoincash-mainnet"
	return &p
}
