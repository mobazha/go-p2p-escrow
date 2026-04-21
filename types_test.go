package escrow

import (
	"encoding/json"
	"testing"
)

func TestAmount_JSON_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		amt  Amount
	}{
		{"zero", NewAmount(0)},
		{"small", NewAmount(100000)},
		{"one_btc", BTC(1.0)},
		{"negative", NewAmount(-42)},
	}

	for _, tt := range tests {
		data, err := json.Marshal(tt.amt)
		if err != nil {
			t.Errorf("%s: Marshal: %v", tt.name, err)
			continue
		}

		var got Amount
		if err := json.Unmarshal(data, &got); err != nil {
			t.Errorf("%s: Unmarshal(%s): %v", tt.name, data, err)
			continue
		}

		if got.Cmp(tt.amt) != 0 {
			t.Errorf("%s: round-trip got %s; want %s", tt.name, got.String(), tt.amt.String())
		}
	}
}

func TestAmount_JSON_InvalidInput(t *testing.T) {
	var a Amount
	if err := json.Unmarshal([]byte(`"not-a-number"`), &a); err == nil {
		t.Error("expected error for invalid amount, got nil")
	}
}

func TestAmount_Arithmetic(t *testing.T) {
	a := NewAmount(100)
	b := NewAmount(50)

	sum := a.Add(b)
	if sum.Int64() != 150 {
		t.Errorf("Add: got %d; want 150", sum.Int64())
	}

	diff := a.Sub(b)
	if diff.Int64() != 50 {
		t.Errorf("Sub: got %d; want 50", diff.Int64())
	}
}

func TestParties_ThresholdAndCount(t *testing.T) {
	p2of2 := Parties{
		Buyer:  Party{PublicKey: []byte{1}},
		Seller: Party{PublicKey: []byte{2}},
	}
	if p2of2.Threshold() != 2 {
		t.Errorf("2-of-2 Threshold = %d; want 2", p2of2.Threshold())
	}
	if p2of2.Count() != 2 {
		t.Errorf("2-of-2 Count = %d; want 2", p2of2.Count())
	}

	mod := Party{PublicKey: []byte{3}}
	p2of3 := Parties{
		Buyer:     Party{PublicKey: []byte{1}},
		Seller:    Party{PublicKey: []byte{2}},
		Moderator: &mod,
	}
	if p2of3.Count() != 3 {
		t.Errorf("2-of-3 Count = %d; want 3", p2of3.Count())
	}
}
