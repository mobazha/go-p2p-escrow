package escrow

import (
	"math/big"
	"time"
)

// Amount represents a monetary value in the chain's smallest unit
// (satoshi, lovelace, wei, piconero, etc.). Backed by [big.Int].
type Amount struct {
	v big.Int
}

// NewAmount creates an Amount from an int64.
func NewAmount(i int64) Amount {
	var a Amount
	a.v.SetInt64(i)
	return a
}

// NewAmountFromBigInt creates an Amount from a [big.Int].
func NewAmountFromBigInt(b *big.Int) Amount {
	var a Amount
	a.v.Set(b)
	return a
}

// BTC returns an Amount given a BTC-denominated float (1 BTC = 1e8 sat).
func BTC(f float64) Amount { return NewAmount(int64(f * 1e8)) }

// Int64 returns the int64 representation. Use [Amount.BigInt] for large values.
func (a Amount) Int64() int64 { return a.v.Int64() }

// BigInt returns a copy of the underlying [big.Int].
func (a Amount) BigInt() *big.Int { return new(big.Int).Set(&a.v) }

// Cmp compares two amounts. Returns -1, 0, or +1.
func (a Amount) Cmp(b Amount) int { return a.v.Cmp(&b.v) }

// Add returns a + b.
func (a Amount) Add(b Amount) Amount {
	var r Amount
	r.v.Add(&a.v, &b.v)
	return r
}

// Sub returns a - b.
func (a Amount) Sub(b Amount) Amount {
	var r Amount
	r.v.Sub(&a.v, &b.v)
	return r
}

// String returns the decimal string representation.
func (a Amount) String() string { return a.v.String() }

// IsZero reports whether the amount is zero.
func (a Amount) IsZero() bool { return a.v.Sign() == 0 }

// SetupParams defines the escrow arrangement to create.
type SetupParams struct {
	Buyer     Party
	Seller    Party
	Moderator *Party        // nil = 2-of-2 (no moderator); non-nil = 2-of-3
	Amount    Amount
	Chain     ChainType
	Timeout   time.Duration // auto-release to seller after timeout; 0 = no timeout
}

// Party represents one participant in an escrow.
type Party struct {
	// PublicKey is the party's public key used for key derivation and signing.
	PublicKey []byte

	// Address is the party's payout address on the target chain.
	// Optional at Setup time; required at Release/Refund time.
	Address string
}

// Parties groups the escrow participants for easy access.
type Parties struct {
	Buyer     Party
	Seller    Party
	Moderator *Party
}

// Threshold returns the number of signatures needed (2 for both 2-of-2 and 2-of-3).
func (p Parties) Threshold() int { return 2 }

// Count returns the total number of parties.
func (p Parties) Count() int {
	if p.Moderator != nil {
		return 3
	}
	return 2
}

// Account represents a created escrow arrangement.
type Account struct {
	// ID is a unique identifier for this escrow (typically derived from the script hash).
	ID string

	// State is the current lifecycle state.
	State EscrowState

	// EscrowAddress is the on-chain address holding the escrowed funds.
	// UTXO: P2WSH address. Monero: multisig subaddress.
	EscrowAddress string

	// Chain identifies which blockchain this escrow lives on.
	Chain ChainType

	// FundingModel describes how the buyer should fund the escrow.
	FundingModel FundingModel

	// Parties are the escrow participants.
	Parties Parties

	// RedeemScript is the chain-specific script data needed to spend.
	// UTXO: the raw redeem script. Other chains: adapter-specific metadata.
	RedeemScript []byte

	// Chaincode is the BIP32 chaincode used for key derivation.
	Chaincode []byte

	// Timeout is the configured auto-release duration. Zero means no timeout.
	Timeout time.Duration

	CreatedAt time.Time
	UpdatedAt time.Time
}

// FundingModel describes how the buyer interacts with the escrow.
type FundingModel string

const (
	// FundingMonitored means the backend monitors the escrow address for incoming funds.
	FundingMonitored FundingModel = "monitored"

	// FundingClientSigned means the frontend builds and signs the funding transaction.
	FundingClientSigned FundingModel = "client_signed"

	// FundingMultiRound means multiple communication rounds are needed (Monero/MPC).
	FundingMultiRound FundingModel = "multi_round"
)

// FundingInstructions tells the buyer how to fund the escrow.
type FundingInstructions struct {
	Address        string // where to send funds
	Amount         Amount // how much to send (including estimated fee for release)
	Memo           string // optional memo/payment ID
	QRData         string // optional QR-encodable payment URI
	ExpiresAt      *time.Time
}

// VerifyParams contains the parameters for verifying escrow funding.
type VerifyParams struct {
	AccountID             string
	RequiredConfirmations int
}

// ReleaseParams contains the parameters for releasing escrowed funds.
type ReleaseParams struct {
	AccountID  string
	ToAddress  string     // payout destination
	Signatures [][]byte   // collected signatures from other parties
}

// RefundParams contains the parameters for refunding escrowed funds.
type RefundParams struct {
	AccountID  string
	ToAddress  string
	Signatures [][]byte
}

// SignParams contains the parameters for signing a release/refund transaction.
type SignParams struct {
	AccountID  string
	PrivateKey []byte // the signer's private key (cleared after use)
	ToAddress  string // where the funds go
}

// Signature is a single party's signature over an escrow transaction input.
type Signature struct {
	InputIndex int
	Data       []byte
}

// ReleaseResult is the outcome of a Release or Refund operation.
type ReleaseResult struct {
	// TxHash is set when the transaction was broadcast (FundingMonitored model).
	TxHash string

	// Signatures are the partial signatures to collect (multi-round models).
	Signatures []Signature
}

// FeeLevel indicates transaction priority.
type FeeLevel int

const (
	FeePriority FeeLevel = iota
	FeeNormal
	FeeEconomic
)

// FeeParams contains the parameters for fee estimation.
type FeeParams struct {
	NumInputs  int
	NumOutputs int
	Level      FeeLevel
}

// AdapterInfo describes an escrow adapter's capabilities.
type AdapterInfo struct {
	Chain          ChainType
	FundingModel   FundingModel
	SupportsTimeout bool
	MinParties     int
	MaxParties     int
}
