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

// TestGetCommittedStateNonExistentAccount verifies that a missing account
// returns the zero hash without creating the account.
func TestGetCommittedStateNonExistentAccount(t *testing.T) {
	db := NewMock()
	addr := types.Address{0x10}
	key := types.Hash{0x01}

	got := db.GetCommittedState(addr, key)
	if got != (types.Hash{}) {
		t.Errorf("expected zero hash for non-existent account, got %x", got)
	}
	if _, exists := db.accounts[addr]; exists {
		t.Error("GetCommittedState must not create accounts as a side effect")
	}
}

// TestGetCommittedStateUnwrittenSlot verifies that for a slot that has never
// been written in this transaction, GetCommittedState returns the current
// storage value (which equals the pre-transaction value).
func TestGetCommittedStateUnwrittenSlot(t *testing.T) {
	db := NewMock()
	addr := types.Address{0x11}
	key := types.Hash{0x02}
	val := types.Hash{0xCC}

	// Pre-populate state as if the slot had a value before this transaction.
	db.accounts[addr] = &Account{
		Balance:        big.NewInt(0),
		State:          map[types.Hash]types.Hash{key: val},
		CommittedState: make(map[types.Hash]types.Hash),
	}

	got := db.GetCommittedState(addr, key)
	if got != val {
		t.Errorf("expected %x (current state) for unwritten slot, got %x", val, got)
	}
}

// TestSetStateSnapshotsOnFirstWrite verifies that the first SetState call for a
// slot captures the pre-transaction value in CommittedState (EIP-2200 v0).
func TestSetStateSnapshotsOnFirstWrite(t *testing.T) {
	db := NewMock()
	addr := types.Address{0x12}
	key := types.Hash{0x03}
	original := types.Hash{0xAA}
	newVal := types.Hash{0xBB}

	// Seed a pre-existing value directly (simulates state at start of tx).
	db.accounts[addr] = &Account{
		Balance:        big.NewInt(0),
		State:          map[types.Hash]types.Hash{key: original},
		CommittedState: make(map[types.Hash]types.Hash),
	}

	db.SetState(addr, key, newVal)

	// Current state must reflect the new value.
	if got := db.GetState(addr, key); got != newVal {
		t.Errorf("GetState: expected %x, got %x", newVal, got)
	}
	// CommittedState must capture the original pre-transaction value.
	if got := db.GetCommittedState(addr, key); got != original {
		t.Errorf("GetCommittedState: expected original %x, got %x", original, got)
	}
}

// TestSetStateDoesNotOverwriteSnapshot verifies that a second write to the same
// slot does not clobber CommittedState — it must keep the very first snapshot.
func TestSetStateDoesNotOverwriteSnapshot(t *testing.T) {
	db := NewMock()
	addr := types.Address{0x13}
	key := types.Hash{0x04}
	original := types.Hash{0x11}
	second := types.Hash{0x22}
	third := types.Hash{0x33}

	db.accounts[addr] = &Account{
		Balance:        big.NewInt(0),
		State:          map[types.Hash]types.Hash{key: original},
		CommittedState: make(map[types.Hash]types.Hash),
	}

	db.SetState(addr, key, second)
	db.SetState(addr, key, third)

	// Current state is the latest write.
	if got := db.GetState(addr, key); got != third {
		t.Errorf("GetState: expected %x, got %x", third, got)
	}
	// CommittedState must still be the original value, not second.
	if got := db.GetCommittedState(addr, key); got != original {
		t.Errorf("GetCommittedState: expected original %x after double write, got %x", original, got)
	}
}

// TestSetStateSnapshotZeroOriginal verifies that a slot that was zero before
// the transaction is snapshotted as zero (not absent).
func TestSetStateSnapshotZeroOriginal(t *testing.T) {
	db := NewMock()
	addr := types.Address{0x14}
	key := types.Hash{0x05}
	newVal := types.Hash{0xFF}

	// Account exists but slot is absent (zero).
	db.SetState(addr, key, newVal)

	if got := db.GetCommittedState(addr, key); got != (types.Hash{}) {
		t.Errorf("expected zero committed state for previously-zero slot, got %x", got)
	}
}

// TestGetCommittedStateIndependentSlots verifies that snapshotting one slot
// does not affect an unrelated slot in the same account.
func TestGetCommittedStateIndependentSlots(t *testing.T) {
	db := NewMock()
	addr := types.Address{0x15}
	keyA := types.Hash{0x0A}
	keyB := types.Hash{0x0B}
	valA := types.Hash{0xA0}
	valB := types.Hash{0xB0}
	newValA := types.Hash{0xA1}

	db.accounts[addr] = &Account{
		Balance:        big.NewInt(0),
		State:          map[types.Hash]types.Hash{keyA: valA, keyB: valB},
		CommittedState: make(map[types.Hash]types.Hash),
	}

	db.SetState(addr, keyA, newValA)

	// Slot A: committed = original valA, current = newValA.
	if got := db.GetCommittedState(addr, keyA); got != valA {
		t.Errorf("slot A committed: expected %x, got %x", valA, got)
	}
	// Slot B was never written — committed falls back to current state (valB).
	if got := db.GetCommittedState(addr, keyB); got != valB {
		t.Errorf("slot B committed: expected %x (unwritten fallback), got %x", valB, got)
	}
}

// TestStateDBCodeSize verifies setting and retrieving code size.
func TestStateDBCodeSize(t *testing.T) {
	db := NewMock()
	addr := types.Address{0x20}

	if got := db.GetCodeSize(addr); got != 0 {
		t.Errorf("expected 0 for non-existent account, got %d", got)
	}

	db.AddCodeSize(addr, 42)
	if got := db.GetCodeSize(addr); got != 42 {
		t.Errorf("expected 42, got %d", got)
	}
}

// TestStateDBCodeHash verifies setting and retrieving the code hash.
func TestStateDBCodeHash(t *testing.T) {
	db := NewMock()
	addr := types.Address{0x21}
	hash := types.Hash{0xDE, 0xAD}

	if got := db.GetCodeHash(addr); got != (types.Hash{}) {
		t.Errorf("expected zero hash for non-existent account, got %x", got)
	}

	db.AddCodeHash(addr, hash)
	if got := db.GetCodeHash(addr); got != hash {
		t.Errorf("expected %x, got %x", hash, got)
	}
}