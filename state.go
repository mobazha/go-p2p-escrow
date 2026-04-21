package escrow

// EscrowState represents a point in the escrow lifecycle.
// It is a named type (not a string alias) so the compiler rejects
// arbitrary strings — only the State* constants are valid.
type EscrowState string

const (
	StateCreated  EscrowState = "created"
	StateFunded   EscrowState = "funded"
	StateReleased EscrowState = "released"
	StateRefunded EscrowState = "refunded"
	StateDisputed EscrowState = "disputed"
	StateResolved EscrowState = "resolved"
	StateExpired  EscrowState = "expired"
	StateSettled  EscrowState = "settled"
)

// StateMachine enforces legal state transitions for an escrow account.
// It prevents invalid operations that could lead to fund loss.
type StateMachine struct {
	transitions map[EscrowState][]EscrowState
}

// NewStateMachine returns a StateMachine with the default transition rules:
//
//	Created  → Funded
//	Funded   → Released, Refunded, Disputed, Expired
//	Released → Settled
//	Refunded → Settled
//	Disputed → Resolved
//	Resolved → Settled
//	Expired  → Settled
func NewStateMachine() *StateMachine {
	return &StateMachine{
		transitions: map[EscrowState][]EscrowState{
			StateCreated:  {StateFunded},
			StateFunded:   {StateReleased, StateRefunded, StateDisputed, StateExpired},
			StateReleased: {StateSettled},
			StateRefunded: {StateSettled},
			StateDisputed: {StateResolved},
			StateResolved: {StateSettled},
			StateExpired:  {StateSettled},
		},
	}
}

// Transition moves account.State from its current value to the target state.
// Returns [ErrInvalidTransition] if the transition is not allowed.
func (sm *StateMachine) Transition(account *Account, to EscrowState) error {
	allowed := sm.transitions[account.State]
	for _, s := range allowed {
		if s == to {
			account.State = to
			return nil
		}
	}
	return ErrInvalidTransition
}

// AllowedTransitions returns the set of states reachable from the given state.
func (sm *StateMachine) AllowedTransitions(from EscrowState) []EscrowState {
	return sm.transitions[from]
}

// IsTerminal reports whether the state is a final state with no further transitions.
func IsTerminal(s EscrowState) bool {
	return s == StateSettled
}
