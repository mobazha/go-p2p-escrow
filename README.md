# go-p2p-escrow

[![CI](https://github.com/mobazha/go-p2p-escrow/actions/workflows/ci.yml/badge.svg)](https://github.com/mobazha/go-p2p-escrow/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/mobazha/go-p2p-escrow.svg)](https://pkg.go.dev/github.com/mobazha/go-p2p-escrow)
[![Go Report Card](https://goreportcard.com/badge/github.com/mobazha/go-p2p-escrow)](https://goreportcard.com/report/github.com/mobazha/go-p2p-escrow)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

**Privacy-preserving, multi-chain P2P escrow SDK for Go.**

Native UTXO multisig, Monero escrow, and multi-chain extensibility — without deploying smart contracts.

> **Status:** v0.1.0 — UTXO escrow for BTC/LTC/BCH/ZEC with P2WSH/P2SH multisig, BIP32 key derivation, and CSV timelock. 43 tests passing.

---

## Why

Every P2P marketplace needs escrow. Building it from scratch means:
- Understanding UTXO scripts, multisig key derivation, and fee estimation per chain
- Implementing a state machine that prevents fund loss
- Handling edge cases (timeouts, disputes, partial signatures)

**go-p2p-escrow** solves this in 20 lines of Go:

```go
package main

import (
    "context"
    "fmt"

    escrow "github.com/mobazha/go-p2p-escrow"
)

func main() {
    ctx := context.Background()
    store := escrow.NewInMemoryStore()
    registry := escrow.NewRegistry(store)

    // Register your chain adapter (see adapters/ for implementations)
    // registry.Register(escrow.ChainBitcoin, utxoAdapter)

    // 1. Create escrow — generates a P2WSH 2-of-3 multisig address
    account, _ := registry.Setup(ctx, escrow.SetupParams{
        Buyer:  escrow.Party{PublicKey: buyerPub},
        Seller: escrow.Party{PublicKey: sellerPub},
        Amount: escrow.BTC(0.01),
        Chain:  escrow.ChainBitcoin,
    })
    fmt.Println("Pay to:", account.EscrowAddress)

    // 2. After buyer funds the address, mark as funded
    registry.MarkFunded(ctx, account.ID, "tx-hash-here")

    // 3. Release funds to seller (or Refund to buyer)
    result, _ := registry.Release(ctx, escrow.ReleaseParams{
        AccountID: account.ID,
        ToAddress: "bc1q-seller-address",
    })
    fmt.Println("Released:", result.TxHash)
}
```

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    go-p2p-escrow                        │
│                                                         │
│  Layer 1: Primitives                                    │
│    ChainType · Amount · Party · ports (KeyProvider,     │
│    UTXOWallet, ChainClient)                             │
│                                                         │
│  Layer 2: Escrow Protocol                               │
│    Escrow interface · Registry · StateMachine ·         │
│    Store · EventHandler · FundingMonitor                │
│                                                         │
│  Layer 3: Chain Adapters                                │
│    adapters/utxo/  (BTC/LTC/BCH/ZEC — P2WSH multisig) │
│    adapters/monero/ (XMR 2-of-3 multisig — planned)    │
│    adapters/evm/    (Safe Module — planned)             │
│                                                         │
│  Layer 4: Transport                                     │
│    libp2p for multi-round protocols (XMR/MPC — planned)│
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### State Machine

The SDK enforces valid state transitions, preventing operations that could lead to fund loss:

```
Created ──→ Funded ──→ Released ──→ Settled
                  │
                  ├──→ Refunded ──→ Settled
                  │
                  ├──→ Disputed ──→ Resolved ──→ Settled
                  │
                  └──→ Expired  ──→ Settled
```

---

## Key Interfaces

| Interface | Purpose | You Implement? |
|---|---|---|
| `Escrow` | Full escrow lifecycle per chain | Only if adding a new chain |
| `Registry` | Dispatches to adapters + orchestrates state/events | No (provided) |
| `StateMachine` | Enforces legal state transitions | No (provided) |
| `Store` | Persist escrow accounts | **Yes** (or use `InMemoryStore` for testing) |
| `EventHandler` | Lifecycle notifications (funded, released, etc.) | **Yes** (or use `NoopEventHandler`) |
| `FundingMonitor` | Watch addresses for incoming payments | Optional |
| `ports.KeyProvider` | Provide master private keys for signing | **Yes** |
| `ports.UTXOWallet` | Low-level UTXO operations (multisig, signing) | **Yes** (for UTXO chains) |

---

## Supported Chains

| Chain | Status | Mechanism | Adapter |
|---|---|---|---|
| Bitcoin (BTC) | ✅ v0.1.0 | P2WSH 2-of-3 multisig + CSV timelock | `adapters/utxo/` |
| Litecoin (LTC) | ✅ v0.1.0 | P2WSH 2-of-3 multisig + CSV timelock | `adapters/utxo/` |
| Bitcoin Cash (BCH) | ✅ v0.1.0 | P2SH 2-of-3 multisig | `adapters/utxo/` |
| Zcash (ZEC) | ✅ v0.1.0 | P2SH 2-of-3 multisig | `adapters/utxo/` |
| Monero (XMR) | 📋 Planned (v0.2) | Native 2-of-3 multisig | `adapters/monero/` |
| Ethereum (ETH) | 📋 Planned (v0.3) | Safe Module | `adapters/evm/safe/` |
| Solana (SOL) | 📋 Planned (v0.3) | Squads Protocol | `adapters/solana/squads/` |

---

## Escrow Lifecycle

```go
// 1. Setup — create escrow, get payment address
account, err := registry.Setup(ctx, escrow.SetupParams{...})

// 2. Fund — buyer sends crypto to account.EscrowAddress
//    SDK monitors or buyer notifies:
registry.MarkFunded(ctx, account.ID, txHash)

// 3a. Release — happy path, buyer confirms delivery
registry.Release(ctx, escrow.ReleaseParams{AccountID: account.ID, ToAddress: sellerAddr})

// 3b. Refund — buyer requests refund, seller agrees
registry.Refund(ctx, escrow.RefundParams{AccountID: account.ID, ToAddress: buyerAddr})

// 3c. Dispute — disagreement, moderator decides
registry.Dispute(ctx, account.ID)
// ... moderator signs Release or Refund ...

// 3d. Expire — timeout, auto-release to seller
registry.MarkExpired(ctx, account.ID)
```

---

## Project Structure

```
go-p2p-escrow/
├── escrow.go          # Escrow interface + Registry (orchestration)
├── types.go           # SetupParams, Account, Party, Amount, FundingModel
├── state.go           # StateMachine (transition enforcement)
├── events.go          # EventHandler interface
├── store.go           # Store interface + InMemoryStore
├── monitor.go         # FundingMonitor interface
├── chain.go           # ChainType constants
├── errors.go          # Sentinel errors
│
├── ports/             # Interfaces for external dependencies
│   ├── keyprovider.go # Master key access
│   ├── wallet.go      # UTXO multisig operations
│   └── chainclient.go # On-chain data (balance, tx info)
│
├── adapters/
│   └── utxo/          # BTC/LTC/BCH/ZEC P2WSH/P2SH multisig + timelock
│       ├── config.go  # Chain presets (BTC/LTC/BCH/ZEC mainnet/testnet/regtest)
│       ├── escrow.go  # Adapter implementing Escrow interface
│       ├── script.go  # M-of-N multisig script builder + address generation
│       ├── timelock.go# CSV timelock scripts (OP_CHECKSEQUENCEVERIFY)
│       └── tx.go      # Transaction building, witness signing, serialization
│
├── crypto/            # BIP32 escrow key derivation
│
└── examples/
    └── btc-escrow/    # Minimal working example
```

---

## Roadmap

| Version | Scope | Grant Window |
|---|---|---|
| **v0.1.0** | UTXO 4-coin P2WSH + state machine + events + example | Zcash retroactive |
| **v0.2.0** | Monero 2-of-3 multisig + libp2p transport | Monero CCS + Zcash FROST |
| **v0.3.0** | EVM (Safe Module) + Solana (Squads) | EF ESP + Safe Foundation |
| **v0.4.0** | MPC/TSS skeleton + FROST POC | libp2p devgrants |

---

## Use Cases

- **P2P marketplaces** — order escrow with buyer protection
- **Telegram Mini App commerce** — in-chat escrow without app downloads
- **OTC trading** — trustless peer-to-peer crypto exchange
- **Freelance platforms** — milestone-based payments
- **AI agent commerce** — autonomous agents trading with escrow guarantees

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines. We use the [Developer Certificate of Origin (DCO)](https://developercertificate.org/).

## Security

If you discover a security vulnerability, please report it responsibly. See [SECURITY.md](SECURITY.md).

## License

[Apache 2.0](LICENSE) — use it freely in commercial and open-source projects.

---

**Built by [Mobazha](https://mobazha.org)** — a privacy-preserving P2P marketplace.
