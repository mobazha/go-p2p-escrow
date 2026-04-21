package escrow

// EventHandler receives notifications about escrow lifecycle events.
// All methods must be idempotent — the SDK may call them more than once
// during retries or recovery.
type EventHandler interface {
	OnStateChanged(account *Account, from, to EscrowState)
	OnFunded(account *Account, txHash string)
	OnReleased(account *Account, result *ReleaseResult)
	OnRefunded(account *Account, result *ReleaseResult)
	OnDisputed(account *Account)
	OnExpired(account *Account)
}

// NoopEventHandler is a default implementation that does nothing.
// Embed it in your handler to only override the events you care about.
type NoopEventHandler struct{}

func (NoopEventHandler) OnStateChanged(*Account, EscrowState, EscrowState) {}
func (NoopEventHandler) OnFunded(*Account, string)                         {}
func (NoopEventHandler) OnReleased(*Account, *ReleaseResult)               {}
func (NoopEventHandler) OnRefunded(*Account, *ReleaseResult)               {}
func (NoopEventHandler) OnDisputed(*Account)                               {}
func (NoopEventHandler) OnExpired(*Account)                                {}
