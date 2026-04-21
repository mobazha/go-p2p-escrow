# go-p2p-escrow

Privacy-preserving, multi-chain P2P escrow SDK for Go.

Native UTXO multisig (BTC/LTC/BCH/ZEC), Monero escrow, and multi-chain extensibility — without smart contracts.

> **Status:** Sprint 0 — API design phase. Interfaces are defined but adapters are not yet implemented.

## Quick Start

```go
store := escrow.NewInMemoryStore()
registry := escrow.NewRegistry(store)
registry.Register(escrow.ChainBitcoin, utxoAdapter)

account, _ := registry.Setup(ctx, escrow.SetupParams{
    Buyer:  escrow.Party{PublicKey: buyerPub},
    Seller: escrow.Party{PublicKey: sellerPub},
    Amount: escrow.BTC(0.01),
    Chain:  escrow.ChainBitcoin,
})

result, _ := registry.Release(ctx, escrow.ReleaseParams{
    AccountID: account.ID,
    ToAddress: sellerAddr,
})
```

## Architecture

```
Layer 1: Primitives     — ChainType, Amount, KeyProvider, UTXOWallet
Layer 2: Escrow Protocol — Escrow interface, Registry, StateMachine, Store, Events
Layer 3: Chain Adapters  — adapters/utxo/, adapters/xmr/ (future)
Layer 4: Transport       — libp2p for multi-round protocols (future)
```

## Key Interfaces

| Interface | Purpose |
|---|---|
| `Escrow` | Full escrow lifecycle (Setup → Fund → Release/Refund) |
| `StateMachine` | Enforces legal state transitions |
| `Store` | Persistence (bring your own DB) |
| `EventHandler` | Lifecycle notifications |
| `FundingMonitor` | Watch escrow addresses for payments |

## Roadmap

| Version | Scope |
|---|---|
| v0.1.0 | UTXO 4-coin P2WSH + state machine + events |
| v0.2.0 | Monero 2-of-3 multisig + transport |
| v0.3.0 | EVM (Safe) + Solana (Squads) |
| v0.4.0 | MPC/TSS skeleton |

## License

Apache 2.0 — see [LICENSE](LICENSE).
