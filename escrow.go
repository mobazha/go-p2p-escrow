package escrow

import (
	"context"
	"fmt"
	"sync"
)

// Escrow manages the full lifecycle of a multi-party escrow arrangement
// on a specific blockchain. Each chain adapter implements this interface.
type Escrow interface {
	// Setup creates a new escrow arrangement and returns an [Account]
	// in [StateCreated]. For UTXO chains this generates a P2WSH multisig
	// address; for Monero it initiates multisig wallet creation.
	Setup(ctx context.Context, params SetupParams) (*Account, error)

	// FundingInfo returns the payment instructions the buyer needs to
	// fund the escrow.
	FundingInfo(account *Account) (*FundingInstructions, error)

	// VerifyFunding checks whether the escrow has been funded with
	// sufficient confirmations. This is a one-shot check; use
	// [FundingMonitor] for continuous monitoring.
	VerifyFunding(ctx context.Context, account *Account, params VerifyParams) (bool, error)

	// Release sends the escrowed funds to the seller (or a specified address).
	// Requires enough signatures to meet the threshold.
	Release(ctx context.Context, account *Account, params ReleaseParams) (*ReleaseResult, error)

	// Refund returns the escrowed funds to the buyer.
	// Requires enough signatures to meet the threshold.
	Refund(ctx context.Context, account *Account, params RefundParams) (*ReleaseResult, error)

	// Sign produces this party's signature(s) for a release or refund
	// transaction. The caller collects signatures from multiple parties
	// and passes them to [Release] or [Refund].
	Sign(ctx context.Context, account *Account, params SignParams) ([]Signature, error)

	// EstimateFee estimates the on-chain fee for an escrow release/refund.
	EstimateFee(ctx context.Context, params FeeParams) (Amount, error)

	// Info returns the adapter's capabilities and metadata.
	Info() AdapterInfo
}

// Registry dispatches escrow operations to the correct chain adapter.
// It also orchestrates state transitions, persistence, and event emission.
type Registry struct {
	mu       sync.RWMutex
	adapters map[ChainType]Escrow
	store    Store
	sm       *StateMachine
	handler  EventHandler
}

// RegistryOption configures a [Registry].
type RegistryOption func(*Registry)

// WithHandler sets the event handler for lifecycle notifications.
func WithHandler(h EventHandler) RegistryOption {
	return func(r *Registry) { r.handler = h }
}

// NewRegistry creates a Registry with the given store and options.
func NewRegistry(store Store, opts ...RegistryOption) *Registry {
	r := &Registry{
		adapters: make(map[ChainType]Escrow),
		store:    store,
		sm:       NewStateMachine(),
		handler:  &NoopEventHandler{},
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Register adds a chain adapter to the registry.
func (r *Registry) Register(chain ChainType, adapter Escrow) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapters[chain] = adapter
}

// Setup creates a new escrow via the appropriate chain adapter,
// persists the account, and emits a state change event.
func (r *Registry) Setup(ctx context.Context, params SetupParams) (*Account, error) {
	adapter, err := r.adapterFor(params.Chain)
	if err != nil {
		return nil, err
	}

	account, err := adapter.Setup(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("setup: %w", err)
	}

	if err := r.store.Save(ctx, account); err != nil {
		return nil, fmt.Errorf("persist: %w", err)
	}

	r.handler.OnStateChanged(account, "", StateCreated)
	return account, nil
}

// Release transitions the escrow to released and sends funds to the seller.
func (r *Registry) Release(ctx context.Context, params ReleaseParams) (*ReleaseResult, error) {
	account, err := r.store.Get(ctx, params.AccountID)
	if err != nil {
		return nil, err
	}

	adapter, err := r.adapterFor(account.Chain)
	if err != nil {
		return nil, err
	}

	prev := account.State
	if err := r.sm.Transition(account, StateReleased); err != nil {
		return nil, err
	}

	result, err := adapter.Release(ctx, account, params)
	if err != nil {
		account.State = prev
		return nil, fmt.Errorf("release: %w", err)
	}

	if err := r.store.UpdateState(ctx, account.ID, StateReleased); err != nil {
		return nil, fmt.Errorf("persist state: %w", err)
	}

	r.handler.OnStateChanged(account, prev, StateReleased)
	r.handler.OnReleased(account, result)
	return result, nil
}

// Refund transitions the escrow to refunded and returns funds to the buyer.
func (r *Registry) Refund(ctx context.Context, params RefundParams) (*ReleaseResult, error) {
	account, err := r.store.Get(ctx, params.AccountID)
	if err != nil {
		return nil, err
	}

	adapter, err := r.adapterFor(account.Chain)
	if err != nil {
		return nil, err
	}

	prev := account.State
	if err := r.sm.Transition(account, StateRefunded); err != nil {
		return nil, err
	}

	result, err := adapter.Refund(ctx, account, params)
	if err != nil {
		account.State = prev
		return nil, fmt.Errorf("refund: %w", err)
	}

	if err := r.store.UpdateState(ctx, account.ID, StateRefunded); err != nil {
		return nil, fmt.Errorf("persist state: %w", err)
	}

	r.handler.OnStateChanged(account, prev, StateRefunded)
	r.handler.OnRefunded(account, result)
	return result, nil
}

func (r *Registry) adapterFor(chain ChainType) (Escrow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.adapters[chain]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedChain, chain)
	}
	return a, nil
}
