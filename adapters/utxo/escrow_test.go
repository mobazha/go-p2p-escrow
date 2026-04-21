package utxo

import (
	"context"
	"testing"
	"time"

	btcec "github.com/btcsuite/btcd/btcec/v2"

	escrow "github.com/mobazha/go-p2p-escrow"
)

func TestAdapter_Setup_2of3(t *testing.T) {
	adapter := NewAdapter(BitcoinRegtest)

	buyer, _ := btcec.NewPrivateKey()
	seller, _ := btcec.NewPrivateKey()
	mod, _ := btcec.NewPrivateKey()

	account, err := adapter.Setup(context.Background(), escrow.SetupParams{
		Buyer:     escrow.Party{PublicKey: buyer.PubKey().SerializeCompressed()},
		Seller:    escrow.Party{PublicKey: seller.PubKey().SerializeCompressed()},
		Moderator: &escrow.Party{PublicKey: mod.PubKey().SerializeCompressed()},
		Chain:     escrow.ChainBitcoin,
	})
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}

	if account.State != escrow.StateCreated {
		t.Errorf("state = %s, want created", account.State)
	}
	if account.EscrowAddress == "" {
		t.Error("empty escrow address")
	}
	if account.EscrowAddress[:5] != "bcrt1" {
		t.Errorf("expected bcrt1 prefix for regtest, got %s", account.EscrowAddress)
	}
	if len(account.RedeemScript) == 0 {
		t.Error("empty redeem script")
	}
	if len(account.Chaincode) != 32 {
		t.Errorf("chaincode length = %d, want 32", len(account.Chaincode))
	}
	if account.FundingModel != escrow.FundingMonitored {
		t.Errorf("funding model = %s, want monitored", account.FundingModel)
	}
}

func TestAdapter_Setup_2of2(t *testing.T) {
	adapter := NewAdapter(BitcoinRegtest)

	buyer, _ := btcec.NewPrivateKey()
	seller, _ := btcec.NewPrivateKey()

	account, err := adapter.Setup(context.Background(), escrow.SetupParams{
		Buyer:  escrow.Party{PublicKey: buyer.PubKey().SerializeCompressed()},
		Seller: escrow.Party{PublicKey: seller.PubKey().SerializeCompressed()},
		Chain:  escrow.ChainBitcoin,
	})
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}

	if account.EscrowAddress == "" {
		t.Error("empty escrow address")
	}
	// 2-of-2 multisig script should not have OP_IF
	if IsTimelockScript(account.RedeemScript) {
		t.Error("2-of-2 should not be a timelock script")
	}
}

func TestAdapter_Setup_WithTimeout(t *testing.T) {
	adapter := NewAdapter(BitcoinRegtest)

	buyer, _ := btcec.NewPrivateKey()
	seller, _ := btcec.NewPrivateKey()
	mod, _ := btcec.NewPrivateKey()

	account, err := adapter.Setup(context.Background(), escrow.SetupParams{
		Buyer:     escrow.Party{PublicKey: buyer.PubKey().SerializeCompressed()},
		Seller:    escrow.Party{PublicKey: seller.PubKey().SerializeCompressed()},
		Moderator: &escrow.Party{PublicKey: mod.PubKey().SerializeCompressed()},
		Chain:     escrow.ChainBitcoin,
		Timeout:   45 * 24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}

	if !IsTimelockScript(account.RedeemScript) {
		t.Error("expected timelock script for non-zero timeout")
	}
	if account.Timeout != 45*24*time.Hour {
		t.Errorf("timeout = %v, want 45 days", account.Timeout)
	}
}

func TestAdapter_Setup_Litecoin(t *testing.T) {
	adapter := NewAdapter(LitecoinMainnet)

	buyer, _ := btcec.NewPrivateKey()
	seller, _ := btcec.NewPrivateKey()
	mod, _ := btcec.NewPrivateKey()

	account, err := adapter.Setup(context.Background(), escrow.SetupParams{
		Buyer:     escrow.Party{PublicKey: buyer.PubKey().SerializeCompressed()},
		Seller:    escrow.Party{PublicKey: seller.PubKey().SerializeCompressed()},
		Moderator: &escrow.Party{PublicKey: mod.PubKey().SerializeCompressed()},
		Chain:     escrow.ChainLitecoin,
	})
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}

	if account.EscrowAddress[:4] != "ltc1" {
		t.Errorf("expected ltc1 prefix, got %s", account.EscrowAddress)
	}
}

func TestAdapter_Setup_ZCash_P2SH(t *testing.T) {
	adapter := NewAdapter(ZCashMainnet)

	buyer, _ := btcec.NewPrivateKey()
	seller, _ := btcec.NewPrivateKey()
	mod, _ := btcec.NewPrivateKey()

	account, err := adapter.Setup(context.Background(), escrow.SetupParams{
		Buyer:     escrow.Party{PublicKey: buyer.PubKey().SerializeCompressed()},
		Seller:    escrow.Party{PublicKey: seller.PubKey().SerializeCompressed()},
		Moderator: &escrow.Party{PublicKey: mod.PubKey().SerializeCompressed()},
		Chain:     escrow.ChainZCash,
	})
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}

	// ZCash uses P2SH, not P2WSH
	if account.EscrowAddress == "" {
		t.Error("empty escrow address")
	}
}

func TestAdapter_Info(t *testing.T) {
	adapter := NewAdapter(BitcoinMainnet)
	info := adapter.Info()

	if info.Chain != escrow.ChainBitcoin {
		t.Errorf("chain = %s, want BTC", info.Chain)
	}
	if info.FundingModel != escrow.FundingMonitored {
		t.Errorf("funding model = %s, want monitored", info.FundingModel)
	}
	if !info.SupportsTimeout {
		t.Error("should support timeout")
	}
	if info.MinParties != 2 || info.MaxParties != 3 {
		t.Errorf("parties = [%d,%d], want [2,3]", info.MinParties, info.MaxParties)
	}
}
