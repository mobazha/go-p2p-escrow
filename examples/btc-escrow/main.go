// Example: Create a Bitcoin P2WSH 2-of-3 multisig escrow.
//
// This demonstrates Setup → address generation in ~20 lines (excluding imports).
// In production, add a ChainClient for on-chain verification and broadcast.
package main

import (
	"context"
	"fmt"
	"log"

	btcec "github.com/btcsuite/btcd/btcec/v2"

	escrow "github.com/mobazha/go-p2p-escrow"
	"github.com/mobazha/go-p2p-escrow/adapters/utxo"
)

func main() {
	ctx := context.Background()
	store := escrow.NewInMemoryStore()
	registry := escrow.NewRegistry(store)

	adapter := utxo.NewAdapter(utxo.BitcoinRegtest)
	registry.Register(escrow.ChainBitcoin, adapter)

	buyer, _ := btcec.NewPrivateKey()
	seller, _ := btcec.NewPrivateKey()
	moderator, _ := btcec.NewPrivateKey()

	account, err := registry.Setup(ctx, escrow.SetupParams{
		Buyer:     escrow.Party{PublicKey: buyer.PubKey().SerializeCompressed()},
		Seller:    escrow.Party{PublicKey: seller.PubKey().SerializeCompressed()},
		Moderator: &escrow.Party{PublicKey: moderator.PubKey().SerializeCompressed()},
		Amount:    escrow.BTC(0.01),
		Chain:     escrow.ChainBitcoin,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Escrow created: %s\n", account.ID)
	fmt.Printf("Pay to (P2WSH): %s\n", account.EscrowAddress)
	fmt.Printf("State: %s\n", account.State)
}
