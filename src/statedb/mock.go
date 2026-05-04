package statedb

import (
	"maps"
	"math/big"

	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

// Account represents an in-memory Ethereum account for the mock DB.
type Account struct {
	Balance        *big.Int
	State          map[types.Hash]types.Hash
	CodeSize       uint64
	CodeHash       types.Hash
	CommittedState map[types.Hash]types.Hash
	Code           []byte
}

// MockStateDB implements types.StateDB for testing and early phases.
type MockStateDB struct {
	accounts  map[types.Address]*Account
	snapshots []map[types.Address]*Account 
}

// NewMock creates a fresh in-memory state database.
func NewMock() *MockStateDB {
	return &MockStateDB{
		accounts: make(map[types.Address]*Account),
	}
}

// getOrCreateAccount retrieves an account, creating it if it doesn't exist.
func (m *MockStateDB) getOrCreateAccount(addr types.Address) *Account {
	acc, exists := m.accounts[addr]
	if !exists {
		acc = &Account{
			Balance:        big.NewInt(0),
			State:          make(map[types.Hash]types.Hash),
			CommittedState: make(map[types.Hash]types.Hash),
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
	return new(big.Int).Set(acc.Balance)
}

// AddBalance adds the specified amount to the account's balance.
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
func (m *MockStateDB) GetState(addr types.Address, key types.Hash) types.Hash {
	acc, exists := m.accounts[addr]
	if !exists {
		return types.Hash{}
	}
	return acc.State[key]
}

// SetState sets a value in the account's storage.
// On first write to a slot, snapshots the pre-transaction value into CommittedState.
func (m *MockStateDB) SetState(addr types.Address, key types.Hash, value types.Hash) {
	acc := m.getOrCreateAccount(addr)
	if _, alreadySnapped := acc.CommittedState[key]; !alreadySnapped {
		acc.CommittedState[key] = acc.State[key]
	}
	acc.State[key] = value
}

// GetCommittedState returns the value of a storage slot as it was at the
// beginning of the current transaction (v0 in EIP-2200's SSTORE gas formula).
func (m *MockStateDB) GetCommittedState(addr types.Address, key types.Hash) types.Hash {
	acc, exists := m.accounts[addr]
	if !exists {
		return types.Hash{}
	}
	if committed, wasSnapped := acc.CommittedState[key]; wasSnapped {
		return committed
	}
	return acc.State[key]
}

// GetCodeSize returns the size of the code associated with the given address.
func (m *MockStateDB) GetCodeSize(addr types.Address) uint64 {
	acc, exists := m.accounts[addr]
	if !exists {
		return 0
	}
	return acc.CodeSize
}

// AddCodeSize sets a mock code size for testing purposes.
func (m *MockStateDB) AddCodeSize(addr types.Address, size uint64) {
	acc := m.getOrCreateAccount(addr)
	acc.CodeSize = size
}

// GetCodeHash returns the Keccak-256 hash of the code associated with the given address.
func (m *MockStateDB) GetCodeHash(addr types.Address) types.Hash {
	acc, exists := m.accounts[addr]
	if !exists {
		return types.Hash{} // Returns empty hash if account doesn't exist
	}
	return acc.CodeHash
}

// AddCodeHash sets a mock code hash for testing purposes.
func (m *MockStateDB) AddCodeHash(addr types.Address, hash types.Hash) {
	acc := m.getOrCreateAccount(addr)
	acc.CodeHash = hash
}

// GetCode returns the bytecode stored at the given address.
func (m *MockStateDB) GetCode(addr types.Address) []byte {
	acc, exists := m.accounts[addr]
	if !exists {
		return nil
	}
	return acc.Code
}

// SetCode stores bytecode at the given address.
func (m *MockStateDB) SetCode(addr types.Address, code []byte) {
	acc := m.getOrCreateAccount(addr)
	acc.Code = code
}

// Snapshot creates a deep copy of the current accounts map, appends it to the
// snapshots slice, and returns the index (id) for later revert.
func (m *MockStateDB) Snapshot() int {
	clone := make(map[types.Address]*Account, len(m.accounts))
	for addr, acc := range m.accounts {
		clonedAcc := &Account{
			Balance:        new(big.Int).Set(acc.Balance),
			CodeSize:       acc.CodeSize,
			CodeHash:       acc.CodeHash,
			Code:           append([]byte{}, acc.Code...),
			State:          make(map[types.Hash]types.Hash, len(acc.State)),
			CommittedState: make(map[types.Hash]types.Hash, len(acc.CommittedState)),
		}
		maps.Copy(clonedAcc.State, acc.State)
		maps.Copy(clonedAcc.CommittedState, acc.CommittedState)
		clone[addr] = clonedAcc
	}
	m.snapshots = append(m.snapshots, clone)
	return len(m.snapshots) - 1
}

// RevertToSnapshot restores the accounts map to the state captured at the given id.
func (m *MockStateDB) RevertToSnapshot(id int) {
	if id < 0 || id >= len(m.snapshots) {
		return
	}
	m.accounts = m.snapshots[id]
}