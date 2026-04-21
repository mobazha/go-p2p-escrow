package utxo

import (
	"testing"
	"time"

	btcec "github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"

	escrow "github.com/mobazha/go-p2p-escrow"
)

func TestSignWitness_Roundtrip(t *testing.T) {
	// Generate 3 keys for 2-of-3 multisig
	privKeys := make([]*btcec.PrivateKey, 3)
	pubKeys := make([]*btcec.PublicKey, 3)
	for i := range privKeys {
		privKeys[i], _ = btcec.NewPrivateKey()
		pubKeys[i] = privKeys[i].PubKey()
	}

	// Create multisig script
	redeemScript, err := CreateMultisigScript(pubKeys, 2)
	if err != nil {
		t.Fatal(err)
	}

	// Create a fake funding UTXO
	fakeHash := chainhash.HashH([]byte("fake-tx"))
	inputAmt := int64(100000) // 0.001 BTC

	// Build unsigned tx paying to a test address
	cfg := &BitcoinRegtest
	inputs := []TxInput{{TxID: fakeHash.String(), Vout: 0, Amount: inputAmt}}
	outputs := []TxOutput{{Address: generateRegtestAddress(t), Amount: 90000}}
	tx, err := BuildUnsignedTx(inputs, outputs, cfg)
	if err != nil {
		t.Fatalf("BuildUnsignedTx: %v", err)
	}

	// Sign with two keys (threshold = 2)
	sigs1, err := SignWitness(tx, privKeys[0], redeemScript, []int64{inputAmt})
	if err != nil {
		t.Fatalf("SignWitness key0: %v", err)
	}
	sigs2, err := SignWitness(tx, privKeys[1], redeemScript, []int64{inputAmt})
	if err != nil {
		t.Fatalf("SignWitness key1: %v", err)
	}

	if len(sigs1) != 1 || len(sigs2) != 1 {
		t.Fatalf("expected 1 sig per key, got %d and %d", len(sigs1), len(sigs2))
	}

	// Assemble the witness
	allSigs := [][]escrow.Signature{{sigs1[0]}, {sigs2[0]}}
	if err := AssembleWitnessMultisig(tx, allSigs, redeemScript, false); err != nil {
		t.Fatalf("AssembleWitnessMultisig: %v", err)
	}

	// Serialize — should not error
	rawTx, err := SerializeTx(tx)
	if err != nil {
		t.Fatalf("SerializeTx: %v", err)
	}
	if len(rawTx) == 0 {
		t.Fatal("empty serialized tx")
	}
}

func TestSignWitness_Timelock_Roundtrip(t *testing.T) {
	privKeys := make([]*btcec.PrivateKey, 3)
	pubKeys := make([]*btcec.PublicKey, 3)
	for i := range privKeys {
		privKeys[i], _ = btcec.NewPrivateKey()
		pubKeys[i] = privKeys[i].PubKey()
	}

	timeoutPriv, _ := btcec.NewPrivateKey()

	// Create timelock script
	redeemScript, err := CreateTimelockScript(pubKeys, 2, 24*time.Hour, timeoutPriv.PubKey())
	if err != nil {
		t.Fatal(err)
	}

	fakeHash := chainhash.HashH([]byte("fake-tx-tl"))
	inputAmt := int64(100000)

	cfg := &BitcoinRegtest
	inputs := []TxInput{{TxID: fakeHash.String(), Vout: 0, Amount: inputAmt}}
	outputs := []TxOutput{{Address: generateRegtestAddress(t), Amount: 90000}}
	tx, err := BuildUnsignedTx(inputs, outputs, cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Multisig path (IF branch): sign with two keys
	sigs1, err := SignWitness(tx, privKeys[0], redeemScript, []int64{inputAmt})
	if err != nil {
		t.Fatal(err)
	}
	sigs2, err := SignWitness(tx, privKeys[1], redeemScript, []int64{inputAmt})
	if err != nil {
		t.Fatal(err)
	}

	allSigs := [][]escrow.Signature{{sigs1[0]}, {sigs2[0]}}
	if err := AssembleWitnessMultisig(tx, allSigs, redeemScript, true); err != nil {
		t.Fatal(err)
	}

	rawTx, err := SerializeTx(tx)
	if err != nil {
		t.Fatal(err)
	}
	if len(rawTx) == 0 {
		t.Fatal("empty serialized tx")
	}
}

func TestSignWitness_TimeoutPath(t *testing.T) {
	privKeys := make([]*btcec.PrivateKey, 3)
	pubKeys := make([]*btcec.PublicKey, 3)
	for i := range privKeys {
		privKeys[i], _ = btcec.NewPrivateKey()
		pubKeys[i] = privKeys[i].PubKey()
	}

	timeoutPriv, _ := btcec.NewPrivateKey()
	timeout := 24 * time.Hour

	redeemScript, err := CreateTimelockScript(pubKeys, 2, timeout, timeoutPriv.PubKey())
	if err != nil {
		t.Fatal(err)
	}

	fakeHash := chainhash.HashH([]byte("fake-tx-timeout"))
	inputAmt := int64(100000)

	cfg := &ChainConfig{
		Chain:       "BTC",
		Params:      &chaincfg.RegressionNetParams,
		AddrType:    P2WSH,
		TxVersion:   1,
		TimelockVer: 2,
	}
	inputs := []TxInput{{TxID: fakeHash.String(), Vout: 0, Amount: inputAmt}}
	outputs := []TxOutput{{Address: generateRegtestAddress(t), Amount: 90000}}
	tx, err := BuildUnsignedTx(inputs, outputs, cfg)
	if err != nil {
		t.Fatal(err)
	}
	tx.Version = cfg.TimelockVer

	seq, err := ParseTimelockSequence(redeemScript)
	if err != nil {
		t.Fatal(err)
	}

	err = AssembleWitnessTimeout(tx, timeoutPriv, redeemScript, []int64{inputAmt}, seq)
	if err != nil {
		t.Fatal(err)
	}

	rawTx, err := SerializeTx(tx)
	if err != nil {
		t.Fatal(err)
	}
	if len(rawTx) == 0 {
		t.Fatal("empty serialized tx")
	}
}

// generateRegtestAddress creates a P2WPKH regtest address for test outputs.
func generateRegtestAddress(t *testing.T) string {
	t.Helper()
	priv, _ := btcec.NewPrivateKey()
	script, _ := CreateMultisigScript([]*btcec.PublicKey{priv.PubKey()}, 1)
	addr, err := ScriptToAddress(script, &chaincfg.RegressionNetParams, P2WSH)
	if err != nil {
		t.Fatalf("generate test address: %v", err)
	}
	return addr
}

