package crypto

import (
	"bytes"
	"testing"

	btcec "github.com/btcsuite/btcd/btcec/v2"
)

func TestDeriveEscrowPublicKey_Deterministic(t *testing.T) {
	priv, _ := btcec.NewPrivateKey()
	chaincode := make([]byte, 32)
	for i := range chaincode {
		chaincode[i] = byte(i)
	}

	pub1, err := DeriveEscrowPublicKey(priv.PubKey(), chaincode)
	if err != nil {
		t.Fatal(err)
	}
	pub2, err := DeriveEscrowPublicKey(priv.PubKey(), chaincode)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(pub1.SerializeCompressed(), pub2.SerializeCompressed()) {
		t.Error("non-deterministic public key derivation")
	}
}

func TestDeriveEscrowPrivateKey_MatchesPublic(t *testing.T) {
	priv, _ := btcec.NewPrivateKey()
	chaincode := make([]byte, 32)
	for i := range chaincode {
		chaincode[i] = byte(i + 42)
	}

	derivedPub, err := DeriveEscrowPublicKey(priv.PubKey(), chaincode)
	if err != nil {
		t.Fatal(err)
	}

	derivedPriv, err := DeriveEscrowPrivateKey(priv, chaincode)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(derivedPriv.PubKey().SerializeCompressed(), derivedPub.SerializeCompressed()) {
		t.Error("derived private key does not match derived public key")
	}
}

func TestDeriveEscrowPublicKey_DifferentChaincodesProduceDifferentKeys(t *testing.T) {
	priv, _ := btcec.NewPrivateKey()

	cc1 := make([]byte, 32)
	cc2 := make([]byte, 32)
	cc2[0] = 1

	pub1, _ := DeriveEscrowPublicKey(priv.PubKey(), cc1)
	pub2, _ := DeriveEscrowPublicKey(priv.PubKey(), cc2)

	if bytes.Equal(pub1.SerializeCompressed(), pub2.SerializeCompressed()) {
		t.Error("different chaincodes should produce different keys")
	}
}
