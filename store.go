package escrow

import (
	"context"
	"sync"
)

// Store persists escrow account state. Consumers provide their own
// implementation (SQLite, PostgreSQL, Redis, etc.). The SDK ships
// [InMemoryStore] for testing.
type Store interface {
	Save(ctx context.Context, account *Account) error
	Get(ctx context.Context, id string) (*Account, error)
	List(ctx context.Context, filter ListFilter) ([]*Account, error)
	UpdateState(ctx context.Context, id string, state EscrowState) error
}

// ListFilter constrains which accounts are returned by [Store.List].
type ListFilter struct {
	Chain  *ChainType
	State  *EscrowState
	Limit  int
	Offset int
}

// InMemoryStore is a thread-safe, in-memory [Store] for testing.
type InMemoryStore struct {
	mu       sync.RWMutex
	accounts map[string]*Account
}

// NewInMemoryStore creates an empty [InMemoryStore].
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{accounts: make(map[string]*Account)}
}

func (s *InMemoryStore) Save(_ context.Context, account *Account) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accounts[account.ID] = deepCopyAccount(account)
	return nil
}

func (s *InMemoryStore) Get(_ context.Context, id string) (*Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.accounts[id]
	if !ok {
		return nil, ErrNotFound
	}
	return deepCopyAccount(a), nil
}

func deepCopyAccount(a *Account) *Account {
	cp := *a
	cp.RedeemScript = copyBytes(a.RedeemScript)
	cp.Chaincode = copyBytes(a.Chaincode)
	cp.Parties.Buyer.PublicKey = copyBytes(a.Parties.Buyer.PublicKey)
	cp.Parties.Seller.PublicKey = copyBytes(a.Parties.Seller.PublicKey)
	if a.Parties.Moderator != nil {
		mod := *a.Parties.Moderator
		mod.PublicKey = copyBytes(a.Parties.Moderator.PublicKey)
		cp.Parties.Moderator = &mod
	}
	return &cp
}

func copyBytes(b []byte) []byte {
	if b == nil {
		return nil
	}
	cp := make([]byte, len(b))
	copy(cp, b)
	return cp
}

func (s *InMemoryStore) List(_ context.Context, filter ListFilter) ([]*Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*Account
	for _, a := range s.accounts {
		if filter.Chain != nil && a.Chain != *filter.Chain {
			continue
		}
		if filter.State != nil && a.State != *filter.State {
			continue
		}
		cp := *a
		result = append(result, &cp)
	}

	start := filter.Offset
	if start > len(result) {
		start = len(result)
	}
	end := len(result)
	if filter.Limit > 0 && start+filter.Limit < end {
		end = start + filter.Limit
	}
	return result[start:end], nil
}

func (s *InMemoryStore) UpdateState(_ context.Context, id string, state EscrowState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.accounts[id]
	if !ok {
		return ErrNotFound
	}
	a.State = state
	return nil
}
