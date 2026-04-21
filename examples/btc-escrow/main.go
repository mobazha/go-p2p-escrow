// Example: Create a Bitcoin escrow and release funds to the seller.
//
// This demonstrates the full lifecycle in ~20 lines (excluding imports).
// In production, replace InMemoryStore and provide a real UTXOWallet.
package main

import (
	"context"
	"fmt"
	"log"

	escrow "github.com/mobazha/go-p2p-escrow"
)

func main() {
	ctx := context.Background()
	store := escrow.NewInMemoryStore()
	registry := escrow.NewRegistry(store)

	// Register your UTXO adapter (see adapters/utxo for the real one)
	// registry.Register(escrow.ChainBitcoin, utxo.New(wallet, keyProvider))

	account, err := registry.Setup(ctx, escrow.SetupParams{
		Buyer:  escrow.Party{PublicKey: []byte("buyer-pub-key")},
		Seller: escrow.Party{PublicKey: []byte("seller-pub-key")},
		Amount: escrow.BTC(0.01),
		Chain:  escrow.ChainBitcoin,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Escrow created: %s at %s\n", account.ID, account.EscrowAddress)

	// After the buyer funds the escrow address, release to the seller:
	result, err := registry.Release(ctx, escrow.ReleaseParams{
		AccountID: account.ID,
		ToAddress: "bc1q-seller-address",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Released: tx %s\n", result.TxHash)
}
