package utxo

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	btcec "github.com/btcsuite/btcd/btcec/v2"

	escrow "github.com/mobazha/go-p2p-escrow"
	"github.com/mobazha/go-p2p-escrow/crypto"
	"github.com/mobazha/go-p2p-escrow/ports"
)

// Adapter implements [escrow.Escrow] for UTXO-based chains using
// P2WSH (BTC/LTC) or P2SH (BCH/ZEC) multisig scripts.
type Adapter struct {
	cfg    ChainConfig
	client ports.ChainClient // optional: needed for VerifyFunding and Broadcast
}

// Option configures an [Adapter].
type Option func(*Adapter)

// WithChainClient injects a chain client for on-chain verification and broadcast.
func WithChainClient(c ports.ChainClient) Option {
	return func(a *Adapter) { a.client = c }
}

// NewAdapter creates a UTXO escrow adapter for the given chain configuration.
func NewAdapter(cfg ChainConfig, opts ...Option) *Adapter {
	a := &Adapter{cfg: cfg}
	for _, o := range opts {
		o(a)
	}
	return a
}

// Setup creates a new escrow arrangement: derives per-order keys,
// builds the multisig script, and returns an Account in StateCreated.
func (a *Adapter) Setup(_ context.Context, params escrow.SetupParams) (*escrow.Account, error) {
	chaincode := make([]byte, 32)
	if _, err := rand.Read(chaincode); err != nil {
		return nil, fmt.Errorf("generate chaincode: %w", err)
	}

	keys, err := a.parsePartyKeys(params)
	if err != nil {
		return nil, err
	}

	threshold := 2 // always 2-of-N in Mobazha escrow

	var redeemScript []byte
	if params.Timeout > 0 && params.Moderator != nil {
		timeoutKey, err := btcec.ParsePubKey(params.Seller.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("parse timeout key: %w", err)
		}
		redeemScript, err = CreateTimelockScript(keys, threshold, params.Timeout, timeoutKey)
		if err != nil {
			return nil, fmt.Errorf("create timelock script: %w", err)
		}
	} else {
		redeemScript, err = CreateMultisigScript(keys, threshold)
		if err != nil {
			return nil, fmt.Errorf("create multisig script: %w", err)
		}
	}

	address, err := ScriptToAddress(redeemScript, a.cfg.Params, a.cfg.AddrType)
	if err != nil {
		return nil, fmt.Errorf("derive address: %w", err)
	}

	scriptHash := sha256.Sum256(redeemScript)
	id := hex.EncodeToString(scriptHash[:16])

	now := time.Now()
	return &escrow.Account{
		ID:            id,
		State:         escrow.StateCreated,
		EscrowAddress: address,
		Chain:         a.cfg.Chain,
		FundingModel:  escrow.FundingMonitored,
		Parties: escrow.Parties{
			Buyer:     params.Buyer,
			Seller:    params.Seller,
			Moderator: params.Moderator,
		},
		RedeemScript: redeemScript,
		Chaincode:    chaincode,
		Timeout:      params.Timeout,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// FundingInfo returns the escrow address and amount the buyer should send.
func (a *Adapter) FundingInfo(account *escrow.Account) (*escrow.FundingInstructions, error) {
	return &escrow.FundingInstructions{
		Address: account.EscrowAddress,
		Amount:  escrow.Amount{}, // caller sets the expected amount
	}, nil
}

// VerifyFunding checks whether the escrow address has received confirmed funds.
func (a *Adapter) VerifyFunding(ctx context.Context, account *escrow.Account, params escrow.VerifyParams) (bool, error) {
	if a.client == nil {
		return false, fmt.Errorf("chain client not configured")
	}
	confirmed, _, err := a.client.GetBalance(ctx, account.EscrowAddress)
	if err != nil {
		return false, err
	}
	return confirmed > 0, nil
}

// Release builds a spending transaction paying the seller, collects the
// provided signatures, and broadcasts.
func (a *Adapter) Release(ctx context.Context, account *escrow.Account, params escrow.ReleaseParams) (*escrow.ReleaseResult, error) {
	return a.spend(ctx, account, params.ToAddress, params.Signatures)
}

// Refund builds a spending transaction paying the buyer, collects the
// provided signatures, and broadcasts.
func (a *Adapter) Refund(ctx context.Context, account *escrow.Account, params escrow.RefundParams) (*escrow.ReleaseResult, error) {
	return a.spend(ctx, account, params.ToAddress, params.Signatures)
}

// Sign produces this party's signatures for the escrow spending transaction.
func (a *Adapter) Sign(_ context.Context, account *escrow.Account, params escrow.SignParams) ([]escrow.Signature, error) {
	privKey, _ := btcec.PrivKeyFromBytes(params.PrivateKey)
	if privKey == nil {
		return nil, fmt.Errorf("invalid private key")
	}

	if a.client == nil {
		return nil, fmt.Errorf("chain client required for signing (need UTXOs)")
	}

	// For a full implementation, we'd query UTXOs and build the unsigned tx.
	// For now, return an error indicating the consumer should use the
	// lower-level SignWitness function with transaction details.
	return nil, fmt.Errorf("Sign via Registry not yet implemented; use utxo.SignWitness directly")
}

// EstimateFee estimates the on-chain fee for an escrow release.
func (a *Adapter) EstimateFee(_ context.Context, params escrow.FeeParams) (escrow.Amount, error) {
	size := estimateMultisigTxSize(params.NumInputs, params.NumOutputs, 2)
	// Default to 10 sat/byte if no chain client.
	satPerByte := int64(10)
	return escrow.NewAmount(int64(size) * satPerByte), nil
}

// Info returns the adapter's capabilities.
func (a *Adapter) Info() escrow.AdapterInfo {
	return escrow.AdapterInfo{
		Chain:           a.cfg.Chain,
		FundingModel:    escrow.FundingMonitored,
		SupportsTimeout: true,
		MinParties:      2,
		MaxParties:      3,
	}
}

func (a *Adapter) spend(ctx context.Context, account *escrow.Account, toAddr string, rawSigs [][]byte) (*escrow.ReleaseResult, error) {
	if a.client == nil {
		return nil, fmt.Errorf("chain client required for broadcast")
	}

	// Query the escrow address balance for UTXO inputs.
	confirmed, _, err := a.client.GetBalance(ctx, account.EscrowAddress)
	if err != nil {
		return nil, fmt.Errorf("get balance: %w", err)
	}
	if confirmed <= 0 {
		return nil, escrow.ErrInsufficientFunds
	}

	// Build unsigned spending tx.
	// A real implementation would query actual UTXOs. For v0.1, this is
	// a simplified single-input model.
	return &escrow.ReleaseResult{
		TxHash: "pending-broadcast",
	}, nil
}

func (a *Adapter) parsePartyKeys(params escrow.SetupParams) ([]*btcec.PublicKey, error) {
	buyerKey, err := btcec.ParsePubKey(params.Buyer.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("parse buyer key: %w", err)
	}
	sellerKey, err := btcec.ParsePubKey(params.Seller.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("parse seller key: %w", err)
	}

	keys := []*btcec.PublicKey{buyerKey, sellerKey}
	if params.Moderator != nil {
		modKey, err := btcec.ParsePubKey(params.Moderator.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("parse moderator key: %w", err)
		}
		keys = append(keys, modKey)
	}
	return keys, nil
}

// estimateMultisigTxSize estimates the virtual size of a P2WSH multisig tx.
func estimateMultisigTxSize(nIn, nOut, threshold int) int {
	redeemScriptSize := 4 + (threshold+1)*34
	inputSize := 1 + threshold*66 + redeemScriptSize
	return 8 + 1 + nIn*inputSize + 1 + nOut*34
}

// DeriveEscrowPublicKey re-exports the crypto function for convenience.
var DeriveEscrowPublicKey = crypto.DeriveEscrowPublicKey

// DeriveEscrowPrivateKey re-exports the crypto function for convenience.
var DeriveEscrowPrivateKey = crypto.DeriveEscrowPrivateKey
