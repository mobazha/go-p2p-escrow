// Package crypto provides BIP32 key derivation for escrow operations.
package crypto

import (
	btcec "github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
)

// DeriveEscrowPublicKey deterministically derives a public key from a master
// public key and a per-order chaincode. Both buyer and seller perform the
// same derivation to arrive at the same escrow key set.
func DeriveEscrowPublicKey(masterPub *btcec.PublicKey, chaincode []byte) (*btcec.PublicKey, error) {
	hdKey := hdkeychain.NewExtendedKey(
		chaincfg.MainNetParams.HDPublicKeyID[:],
		masterPub.SerializeCompressed(),
		chaincode,
		[]byte{0x00, 0x00, 0x00, 0x00},
		0, 0, false,
	)
	child, err := deriveFirstValidChild(hdKey)
	if err != nil {
		return nil, err
	}
	return child.ECPubKey()
}

// DeriveEscrowPrivateKey deterministically derives a private key from a master
// private key and a per-order chaincode. Used when signing a release/refund.
func DeriveEscrowPrivateKey(masterPriv *btcec.PrivateKey, chaincode []byte) (*btcec.PrivateKey, error) {
	hdKey := hdkeychain.NewExtendedKey(
		chaincfg.MainNetParams.HDPrivateKeyID[:],
		masterPriv.Serialize(),
		chaincode,
		[]byte{0x00, 0x00, 0x00, 0x00},
		0, 0, true,
	)
	child, err := deriveFirstValidChild(hdKey)
	if err != nil {
		return nil, err
	}
	return child.ECPrivKey()
}

// deriveFirstValidChild derives child 0. If child 0 is invalid (rare edge
// case in BIP32), it increments until a valid child is found.
func deriveFirstValidChild(parent *hdkeychain.ExtendedKey) (*hdkeychain.ExtendedKey, error) {
	for i := uint32(0); ; i++ {
		child, err := parent.Derive(i)
		if err != nil {
			continue
		}
		return child, nil
	}
}
