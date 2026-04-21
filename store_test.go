package escrow

import (
	"context"
	"testing"
)

func TestInMemoryStore_SaveAndGet(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()

	account := &Account{
		ID:    "test-1",
		State: StateCreated,
		Chain: ChainBitcoin,
	}

	if err := store.Save(ctx, account); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := store.Get(ctx, "test-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != "test-1" || got.State != StateCreated {
		t.Errorf("Get = {%s, %s}; want {test-1, created}", got.ID, got.State)
	}
}

func TestInMemoryStore_GetNotFound(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()

	_, err := store.Get(ctx, "nonexistent")
	if err != ErrNotFound {
		t.Errorf("Get(nonexistent) = %v; want ErrNotFound", err)
	}
}

func TestInMemoryStore_UpdateState(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()

	store.Save(ctx, &Account{ID: "test-2", State: StateCreated, Chain: ChainBitcoin})

	if err := store.UpdateState(ctx, "test-2", StateFunded); err != nil {
		t.Fatalf("UpdateState: %v", err)
	}

	got, _ := store.Get(ctx, "test-2")
	if got.State != StateFunded {
		t.Errorf("State = %s; want funded", got.State)
	}
}

func TestInMemoryStore_List(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()

	btc := ChainBitcoin
	ltc := ChainLitecoin
	funded := StateFunded

	store.Save(ctx, &Account{ID: "a1", State: StateCreated, Chain: ChainBitcoin})
	store.Save(ctx, &Account{ID: "a2", State: StateFunded, Chain: ChainBitcoin})
	store.Save(ctx, &Account{ID: "a3", State: StateFunded, Chain: ChainLitecoin})

	tests := []struct {
		name   string
		filter ListFilter
		want   int
	}{
		{"all", ListFilter{}, 3},
		{"btc only", ListFilter{Chain: &btc}, 2},
		{"ltc only", ListFilter{Chain: &ltc}, 1},
		{"funded only", ListFilter{State: &funded}, 2},
		{"btc+funded", ListFilter{Chain: &btc, State: &funded}, 1},
		{"limit 1", ListFilter{Limit: 1}, 1},
	}

	for _, tt := range tests {
		got, err := store.List(ctx, tt.filter)
		if err != nil {
			t.Errorf("%s: List: %v", tt.name, err)
			continue
		}
		if len(got) != tt.want {
			t.Errorf("%s: len = %d; want %d", tt.name, len(got), tt.want)
		}
	}
}

func TestInMemoryStore_IsolatesMutations(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()

	original := &Account{ID: "iso-1", State: StateCreated, Chain: ChainBitcoin}
	store.Save(ctx, original)

	got, _ := store.Get(ctx, "iso-1")
	got.State = StateFunded

	check, _ := store.Get(ctx, "iso-1")
	if check.State != StateCreated {
		t.Error("Store returned a reference instead of a copy; mutations leak")
	}
}

func TestInMemoryStore_DeepCopyBytes(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()

	original := &Account{
		ID:           "deep-1",
		State:        StateCreated,
		Chain:        ChainBitcoin,
		RedeemScript: []byte{0x01, 0x02, 0x03},
		Parties: Parties{
			Buyer:  Party{PublicKey: []byte{0xaa, 0xbb}},
			Seller: Party{PublicKey: []byte{0xcc, 0xdd}},
		},
	}
	store.Save(ctx, original)

	got, _ := store.Get(ctx, "deep-1")
	got.RedeemScript[0] = 0xff
	got.Parties.Buyer.PublicKey[0] = 0xff

	check, _ := store.Get(ctx, "deep-1")
	if check.RedeemScript[0] != 0x01 {
		t.Error("RedeemScript byte slice was not deep-copied")
	}
	if check.Parties.Buyer.PublicKey[0] != 0xaa {
		t.Error("Buyer.PublicKey byte slice was not deep-copied")
	}
}
