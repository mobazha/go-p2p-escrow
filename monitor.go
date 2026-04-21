package escrow

import (
	"context"
	"time"
)

// FundingMonitor watches escrow addresses for incoming payments and
// triggers callbacks when funding is confirmed or times out.
// UTXO chains typically poll an Electrum or mempool backend.
type FundingMonitor interface {
	// Watch begins monitoring the escrow address for incoming funds.
	// It calls opts.OnConfirmed when the required confirmations are reached,
	// or opts.OnTimeout if the deadline expires.
	Watch(ctx context.Context, account *Account, opts WatchOpts) error

	// Stop cancels monitoring for the given account.
	Stop(accountID string) error
}

// WatchOpts configures a [FundingMonitor.Watch] call.
type WatchOpts struct {
	// RequiredConfirmations is the number of block confirmations needed.
	RequiredConfirmations int

	// Timeout is how long to wait before giving up. Zero means no timeout.
	Timeout time.Duration

	// OnConfirmed is called when the payment reaches the required confirmations.
	OnConfirmed func(account *Account, txHash string)

	// OnTimeout is called if the payment is not received in time.
	OnTimeout func(account *Account)
}
