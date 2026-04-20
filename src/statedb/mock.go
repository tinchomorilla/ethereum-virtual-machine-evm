package statedb

import (
	"math/big"

	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

// Account represents an in-memory Ethereum account for the mock DB.
type Account struct {
	Balance *big.Int
	State   map[types.Hash]types.Hash
}

// MockStateDB implements types.StateDB for testing and early phases.
type MockStateDB struct {
	accounts map[types.Address]*Account
}

// NewMock creates a fresh in-memory state database.
func NewMock() *MockStateDB {
	return &MockStateDB{
		accounts: make(map[types.Address]*Account),
	}
}

// getOrCreateAccount retrieves an account, creating it if it doesn't exist (for write paths).
func (m *MockStateDB) getOrCreateAccount(addr types.Address) *Account {
	acc, exists := m.accounts[addr]
	if !exists {
		acc = &Account{
			Balance: big.NewInt(0),
			State:   make(map[types.Hash]types.Hash),
		}
		m.accounts[addr] = acc
	}
	return acc
}

// GetBalance returns the balance of the given address, or 0 if it doesn't exist.
func (m *MockStateDB) GetBalance(addr types.Address) *big.Int {
	acc, exists := m.accounts[addr]
	if !exists {
		return big.NewInt(0)
	}
	// Creates an independent copy, so the caller can do whatever they want with it without corrupting the DB's state.
	return new(big.Int).Set(acc.Balance)
}

// AddBalance adds the specified amount to the account's balance.
// If the account doesn't exist, it will be created with the given balance.
func (m *MockStateDB) AddBalance(addr types.Address, amount *big.Int) {
	if amount.Sign() < 0 {
		panic("AddBalance: negative amount")
	}
	acc := m.getOrCreateAccount(addr)
	acc.Balance.Add(acc.Balance, amount)
}

// SubBalance subtracts the specified amount from the account's balance.
func (m *MockStateDB) SubBalance(addr types.Address, amount *big.Int) {
	acc := m.getOrCreateAccount(addr)
	acc.Balance.Sub(acc.Balance, amount)
}

// GetState retrieves a value from the account's storage.
// Returns zero hash for non-existent accounts without creating them (read-only, no side effects).
func (m *MockStateDB) GetState(addr types.Address, key types.Hash) types.Hash {
	acc, exists := m.accounts[addr]
	if !exists {
		return types.Hash{}
	}
	return acc.State[key]
}

// SetState sets a value in the account's storage.
func (m *MockStateDB) SetState(addr types.Address, key types.Hash, value types.Hash) {
	acc := m.getOrCreateAccount(addr)
	acc.State[key] = value
}