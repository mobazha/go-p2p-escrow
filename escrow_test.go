package escrow

import (
	"context"
	"testing"
)

type stubAdapter struct {
	setupCalled  bool
	signCalled   bool
	releaseCalled bool
}

func (s *stubAdapter) Setup(_ context.Context, params SetupParams) (*Account, error) {
	s.setupCalled = true
	return &Account{
		ID:            "stub-1",
		State:         StateCreated,
		EscrowAddress: "bc1q-stub",
		Chain:         params.Chain,
		Parties: Parties{
			Buyer:  params.Buyer,
			Seller: params.Seller,
		},
	}, nil
}

func (s *stubAdapter) FundingInfo(account *Account) (*FundingInstructions, error) {
	return &FundingInstructions{Address: account.EscrowAddress, Amount: BTC(0.01)}, nil
}

func (s *stubAdapter) VerifyFunding(_ context.Context, _ *Account, _ VerifyParams) (bool, error) {
	return true, nil
}

func (s *stubAdapter) Release(_ context.Context, _ *Account, _ ReleaseParams) (*ReleaseResult, error) {
	s.releaseCalled = true
	return &ReleaseResult{TxHash: "tx-release"}, nil
}

func (s *stubAdapter) Refund(_ context.Context, _ *Account, _ RefundParams) (*ReleaseResult, error) {
	return &ReleaseResult{TxHash: "tx-refund"}, nil
}

func (s *stubAdapter) Sign(_ context.Context, _ *Account, _ SignParams) ([]Signature, error) {
	s.signCalled = true
	return []Signature{{InputIndex: 0, Data: []byte("sig")}}, nil
}

func (s *stubAdapter) EstimateFee(_ context.Context, _ FeeParams) (Amount, error) {
	return NewAmount(1000), nil
}

func (s *stubAdapter) Info() AdapterInfo {
	return AdapterInfo{Chain: ChainBitcoin, FundingModel: FundingMonitored}
}

func TestRegistry_FullLifecycle(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()
	adapter := &stubAdapter{}

	var events []EscrowState
	handler := &lifecycleHandler{onState: func(a *Account, from, to EscrowState) {
		events = append(events, to)
	}}

	reg := NewRegistry(store, WithHandler(handler))
	reg.Register(ChainBitcoin, adapter)

	account, err := reg.Setup(ctx, SetupParams{
		Buyer:  Party{PublicKey: []byte("buyer")},
		Seller: Party{PublicKey: []byte("seller")},
		Amount: BTC(0.01),
		Chain:  ChainBitcoin,
	})
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}
	if account.State != StateCreated {
		t.Fatalf("State = %s; want created", account.State)
	}

	if err := reg.MarkFunded(ctx, account.ID, "tx-fund"); err != nil {
		t.Fatalf("MarkFunded: %v", err)
	}

	sigs, err := reg.Sign(ctx, SignParams{AccountID: account.ID, PrivateKey: []byte("key"), ToAddress: "seller-addr"})
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if len(sigs) != 1 {
		t.Errorf("Sign returned %d sigs; want 1", len(sigs))
	}

	result, err := reg.Release(ctx, ReleaseParams{AccountID: account.ID, ToAddress: "seller-addr"})
	if err != nil {
		t.Fatalf("Release: %v", err)
	}
	if result.TxHash != "tx-release" {
		t.Errorf("TxHash = %s; want tx-release", result.TxHash)
	}

	got, _ := reg.Get(ctx, account.ID)
	if got.State != StateReleased {
		t.Errorf("Final state = %s; want released", got.State)
	}

	if len(events) != 3 {
		t.Errorf("events = %d; want 3 (created, funded, released)", len(events))
	}
}

func TestRegistry_DisputeFlow(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()
	adapter := &stubAdapter{}

	reg := NewRegistry(store)
	reg.Register(ChainBitcoin, adapter)

	account, _ := reg.Setup(ctx, SetupParams{
		Buyer:  Party{PublicKey: []byte("buyer")},
		Seller: Party{PublicKey: []byte("seller")},
		Amount: BTC(0.05),
		Chain:  ChainBitcoin,
	})

	reg.MarkFunded(ctx, account.ID, "tx-fund")

	if err := reg.Dispute(ctx, account.ID); err != nil {
		t.Fatalf("Dispute: %v", err)
	}

	got, _ := reg.Get(ctx, account.ID)
	if got.State != StateDisputed {
		t.Errorf("State = %s; want disputed", got.State)
	}
}

func TestRegistry_InvalidTransition(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()
	adapter := &stubAdapter{}

	reg := NewRegistry(store)
	reg.Register(ChainBitcoin, adapter)

	account, _ := reg.Setup(ctx, SetupParams{
		Buyer:  Party{PublicKey: []byte("buyer")},
		Seller: Party{PublicKey: []byte("seller")},
		Amount: BTC(0.01),
		Chain:  ChainBitcoin,
	})

	_, err := reg.Release(ctx, ReleaseParams{AccountID: account.ID, ToAddress: "seller"})
	if err != ErrInvalidTransition {
		t.Errorf("Release from created = %v; want ErrInvalidTransition", err)
	}
}

func TestRegistry_UnsupportedChain(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()
	reg := NewRegistry(store)

	_, err := reg.Setup(ctx, SetupParams{Chain: ChainMonero})
	if err == nil {
		t.Error("Setup with no adapter registered should fail")
	}
}

type lifecycleHandler struct {
	NoopEventHandler
	onState func(*Account, EscrowState, EscrowState)
}

func (h *lifecycleHandler) OnStateChanged(a *Account, from, to EscrowState) {
	if h.onState != nil {
		h.onState(a, from, to)
	}
}
