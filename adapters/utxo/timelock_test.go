package utxo

import (
	"testing"
	"time"

	btcec "github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

func TestCreateTimelockScript(t *testing.T) {
	keys := generateTestKeys(3)
	timeoutKey, _ := btcec.NewPrivateKey()

	script, err := CreateTimelockScript(keys, 2, 24*time.Hour, timeoutKey.PubKey())
	if err != nil {
		t.Fatalf("CreateTimelockScript: %v", err)
	}

	// Should start with OP_IF
	if script[0] != txscript.OP_IF {
		t.Errorf("script[0] = 0x%02x, want 0x63 (OP_IF)", script[0])
	}

	// Should end with OP_ENDIF
	if script[len(script)-1] != txscript.OP_ENDIF {
		t.Errorf("script[-1] = 0x%02x, want 0x68 (OP_ENDIF)", script[len(script)-1])
	}
}

func TestTimelockScript_ToAddress(t *testing.T) {
	keys := generateTestKeys(3)
	timeoutKey, _ := btcec.NewPrivateKey()

	script, err := CreateTimelockScript(keys, 2, 48*time.Hour, timeoutKey.PubKey())
	if err != nil {
		t.Fatal(err)
	}

	addr, err := ScriptToAddress(script, &chaincfg.MainNetParams, P2WSH)
	if err != nil {
		t.Fatalf("ScriptToAddress: %v", err)
	}
	if addr[:4] != "bc1q" {
		t.Errorf("expected bc1q prefix, got %s", addr)
	}
}

func TestIsTimelockScript(t *testing.T) {
	keys := generateTestKeys(3)

	// Regular multisig
	regular, _ := CreateMultisigScript(keys, 2)
	if IsTimelockScript(regular) {
		t.Error("regular script should not be timelock")
	}

	// Timelock
	timeoutKey, _ := btcec.NewPrivateKey()
	tl, _ := CreateTimelockScript(keys, 2, time.Hour, timeoutKey.PubKey())
	if !IsTimelockScript(tl) {
		t.Error("timelock script should be detected")
	}
}

func TestParseTimelockSequence(t *testing.T) {
	keys := generateTestKeys(3)
	timeoutKey, _ := btcec.NewPrivateKey()

	// 24 hours ≈ 144 blocks (6 blocks/hour)
	script, err := CreateTimelockScript(keys, 2, 24*time.Hour, timeoutKey.PubKey())
	if err != nil {
		t.Fatal(err)
	}

	seq, err := ParseTimelockSequence(script)
	if err != nil {
		t.Fatalf("ParseTimelockSequence: %v", err)
	}
	if seq == 0 {
		t.Error("expected non-zero sequence")
	}
}
