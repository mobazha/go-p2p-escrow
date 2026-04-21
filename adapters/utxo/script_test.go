package utxo

import (
	"testing"

	btcec "github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
)

func generateTestKeys(n int) []*btcec.PublicKey {
	keys := make([]*btcec.PublicKey, n)
	for i := range keys {
		priv, _ := btcec.NewPrivateKey()
		keys[i] = priv.PubKey()
	}
	return keys
}

func TestCreateMultisigScript_2of3(t *testing.T) {
	keys := generateTestKeys(3)
	script, err := CreateMultisigScript(keys, 2)
	if err != nil {
		t.Fatalf("CreateMultisigScript: %v", err)
	}
	if len(script) == 0 {
		t.Fatal("empty script")
	}
	// Script should start with OP_2 (0x52)
	if script[0] != 0x52 {
		t.Errorf("script[0] = 0x%02x, want 0x52 (OP_2)", script[0])
	}
}

func TestCreateMultisigScript_1of2(t *testing.T) {
	keys := generateTestKeys(2)
	script, err := CreateMultisigScript(keys, 1)
	if err != nil {
		t.Fatalf("CreateMultisigScript: %v", err)
	}
	// Script should start with OP_1 (0x51)
	if script[0] != 0x51 {
		t.Errorf("script[0] = 0x%02x, want 0x51 (OP_1)", script[0])
	}
}

func TestCreateMultisigScript_TooFewKeys(t *testing.T) {
	keys := generateTestKeys(1)
	_, err := CreateMultisigScript(keys, 2)
	if err == nil {
		t.Error("expected error for threshold > key count")
	}
}

func TestCreateMultisigScript_TooManyKeys(t *testing.T) {
	keys := generateTestKeys(9)
	_, err := CreateMultisigScript(keys, 2)
	if err == nil {
		t.Error("expected error for >8 keys")
	}
}

func TestScriptToAddress_P2WSH_Bitcoin(t *testing.T) {
	keys := generateTestKeys(3)
	script, err := CreateMultisigScript(keys, 2)
	if err != nil {
		t.Fatalf("CreateMultisigScript: %v", err)
	}

	addr, err := ScriptToAddress(script, &chaincfg.MainNetParams, P2WSH)
	if err != nil {
		t.Fatalf("ScriptToAddress: %v", err)
	}
	if addr[:4] != "bc1q" {
		t.Errorf("expected bc1q prefix, got %s", addr)
	}
}

func TestScriptToAddress_P2WSH_Testnet(t *testing.T) {
	keys := generateTestKeys(3)
	script, err := CreateMultisigScript(keys, 2)
	if err != nil {
		t.Fatalf("CreateMultisigScript: %v", err)
	}

	addr, err := ScriptToAddress(script, &chaincfg.TestNet3Params, P2WSH)
	if err != nil {
		t.Fatalf("ScriptToAddress: %v", err)
	}
	if addr[:4] != "tb1q" {
		t.Errorf("expected tb1q prefix, got %s", addr)
	}
}

func TestScriptToAddress_P2WSH_Litecoin(t *testing.T) {
	keys := generateTestKeys(3)
	script, err := CreateMultisigScript(keys, 2)
	if err != nil {
		t.Fatalf("CreateMultisigScript: %v", err)
	}

	addr, err := ScriptToAddress(script, litecoinMainNetParams(), P2WSH)
	if err != nil {
		t.Fatalf("ScriptToAddress: %v", err)
	}
	if addr[:4] != "ltc1" {
		t.Errorf("expected ltc1 prefix, got %s", addr)
	}
}

func TestScriptToAddress_P2SH(t *testing.T) {
	keys := generateTestKeys(3)
	script, err := CreateMultisigScript(keys, 2)
	if err != nil {
		t.Fatalf("CreateMultisigScript: %v", err)
	}

	addr, err := ScriptToAddress(script, &chaincfg.MainNetParams, P2SH)
	if err != nil {
		t.Fatalf("ScriptToAddress: %v", err)
	}
	if addr[0] != '3' {
		t.Errorf("expected P2SH address starting with 3, got %s", addr)
	}
}

func TestScriptToAddress_Deterministic(t *testing.T) {
	keys := generateTestKeys(3)
	script, err := CreateMultisigScript(keys, 2)
	if err != nil {
		t.Fatal(err)
	}

	addr1, _ := ScriptToAddress(script, &chaincfg.MainNetParams, P2WSH)
	addr2, _ := ScriptToAddress(script, &chaincfg.MainNetParams, P2WSH)
	if addr1 != addr2 {
		t.Errorf("non-deterministic: %s != %s", addr1, addr2)
	}
}
