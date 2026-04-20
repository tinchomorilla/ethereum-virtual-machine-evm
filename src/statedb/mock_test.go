package statedb

import (
	"math/big"
	"testing"

	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

// TestStateDBBalance verifies the logic for adding, subtracting and reading balances.
func TestStateDBBalance(t *testing.T) {
	db := NewMock()
	addr := types.Address{0x01}

	// 1. Unused accounts should return 0, not panic
	bal := db.GetBalance(addr)
	if bal.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("expected 0 balance for new account, got %v", bal)
	}

	// 2. Add balance
	db.AddBalance(addr, big.NewInt(100))
	if db.GetBalance(addr).Cmp(big.NewInt(100)) != 0 {
		t.Errorf("expected 100 after addition")
	}

	// 3. Sub balance
	db.SubBalance(addr, big.NewInt(30))
	if db.GetBalance(addr).Cmp(big.NewInt(70)) != 0 {
		t.Errorf("expected 70 after subtraction")
	}
}

// TestStateDBPointerIsolation verifies that modifying the returned balance 
// doesn't corrupt the actual database state.
func TestStateDBPointerIsolation(t *testing.T) {
	db := NewMock()
	addr := types.Address{0x02}

	db.AddBalance(addr, big.NewInt(50))
	
	// Retrieve the balance and maliciously modify the returned pointer
	bal := db.GetBalance(addr)
	bal.Add(bal, big.NewInt(999))

	// The actual database should remain untouched
	actualBal := db.GetBalance(addr)
	if actualBal.Cmp(big.NewInt(50)) != 0 {
		t.Errorf("DB state corrupted! Expected 50, got %v", actualBal)
	}
}

// TestStateDBStorage verifies the key-value storage at the account level.
func TestStateDBStorage(t *testing.T) {
	db := NewMock()
	addr := types.Address{0x03}
	key := types.Hash{0xAA}
	val := types.Hash{0xBB}

	// 1. Empty storage should return a zero-hash
	emptyHash := db.GetState(addr, key)
	if emptyHash != (types.Hash{}) {
		t.Errorf("expected empty hash, got %x", emptyHash)
	}

	// 2. Set and Get
	db.SetState(addr, key, val)
	retrieved := db.GetState(addr, key)
	if retrieved != val {
		t.Errorf("expected %x, got %x", val, retrieved)
	}
}