package escrow

import "testing"

func TestStateMachine_ValidTransitions(t *testing.T) {
	sm := NewStateMachine()

	tests := []struct {
		from, to EscrowState
	}{
		{StateCreated, StateFunded},
		{StateFunded, StateReleased},
		{StateFunded, StateRefunded},
		{StateFunded, StateDisputed},
		{StateFunded, StateExpired},
		{StateReleased, StateSettled},
		{StateRefunded, StateSettled},
		{StateDisputed, StateResolved},
		{StateResolved, StateSettled},
		{StateExpired, StateSettled},
	}

	for _, tt := range tests {
		account := &Account{State: tt.from}
		if err := sm.Transition(account, tt.to); err != nil {
			t.Errorf("Transition(%s → %s) = %v; want nil", tt.from, tt.to, err)
		}
		if account.State != tt.to {
			t.Errorf("account.State = %s; want %s", account.State, tt.to)
		}
	}
}

func TestStateMachine_InvalidTransitions(t *testing.T) {
	sm := NewStateMachine()

	tests := []struct {
		from, to EscrowState
	}{
		{StateCreated, StateReleased},
		{StateCreated, StateSettled},
		{StateFunded, StateCreated},
		{StateFunded, StateSettled},
		{StateReleased, StateFunded},
		{StateSettled, StateCreated},
		{StateDisputed, StateReleased},
	}

	for _, tt := range tests {
		account := &Account{State: tt.from}
		if err := sm.Transition(account, tt.to); err != ErrInvalidTransition {
			t.Errorf("Transition(%s → %s) = %v; want ErrInvalidTransition", tt.from, tt.to, err)
		}
		if account.State != tt.from {
			t.Errorf("account.State changed to %s after invalid transition; should stay %s", account.State, tt.from)
		}
	}
}

func TestStateMachine_AllowedTransitions(t *testing.T) {
	sm := NewStateMachine()

	allowed := sm.AllowedTransitions(StateFunded)
	if len(allowed) != 4 {
		t.Errorf("AllowedTransitions(funded) = %d items; want 4", len(allowed))
	}

	allowed = sm.AllowedTransitions(StateSettled)
	if len(allowed) != 0 {
		t.Errorf("AllowedTransitions(settled) = %d items; want 0", len(allowed))
	}
}

func TestIsTerminal(t *testing.T) {
	if !IsTerminal(StateSettled) {
		t.Error("IsTerminal(settled) = false; want true")
	}
	if IsTerminal(StateFunded) {
		t.Error("IsTerminal(funded) = true; want false")
	}
}
