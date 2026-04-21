package escrow

import "errors"

var (
	// ErrInvalidTransition is returned when a state transition is not allowed
	// by the [StateMachine]. This prevents operations that could lead to fund loss.
	ErrInvalidTransition = errors.New("escrow: invalid state transition")

	// ErrNotFound is returned when an escrow account does not exist in the [Store].
	ErrNotFound = errors.New("escrow: account not found")

	// ErrInsufficientFunds is returned when the escrow address does not hold
	// enough confirmed funds.
	ErrInsufficientFunds = errors.New("escrow: insufficient funds")

	// ErrAlreadyFunded is returned when attempting to fund an escrow that
	// has already been funded.
	ErrAlreadyFunded = errors.New("escrow: already funded")

	// ErrUnsupportedChain is returned when no adapter is registered for
	// the requested [ChainType].
	ErrUnsupportedChain = errors.New("escrow: unsupported chain")

	// ErrThresholdNotMet is returned when the number of valid signatures
	// is less than the escrow threshold.
	ErrThresholdNotMet = errors.New("escrow: signature threshold not met")

	// ErrTimeout is returned when the escrow has expired.
	ErrTimeout = errors.New("escrow: timeout expired")
)
