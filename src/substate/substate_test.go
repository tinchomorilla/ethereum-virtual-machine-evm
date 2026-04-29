package substate

import (
	"testing"

	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

// TestNewAccruedSubstateZeroState verifies that the fresh substate starts empty
// and with a zero refund counter.
func TestNewAccruedSubstateZeroState(t *testing.T) {
	a := NewAccruedSubstate()

	if len(a.SelfDestructs) != 0 {
		t.Errorf("expected empty SelfDestructs, got %d entries", len(a.SelfDestructs))
	}
	if len(a.Logs) != 0 {
		t.Errorf("expected no logs, got %d", len(a.Logs))
	}
	if len(a.TouchedAccounts) != 0 {
		t.Errorf("expected empty TouchedAccounts, got %d entries", len(a.TouchedAccounts))
	}
	if a.Refund != 0 {
		t.Errorf("expected zero refund, got %d", a.Refund)
	}
	if len(a.WarmStorage) != 0 {
		t.Errorf("expected empty WarmStorage, got %d entries", len(a.WarmStorage))
	}
}

// TestNewAccruedSubstatePrecompiledWarm verifies that the 9 precompiled
// addresses (0x01..0x09) are warm from the start, as required by EIP-2929.
func TestNewAccruedSubstatePrecompiledWarm(t *testing.T) {
	a := NewAccruedSubstate()

	for i := 1; i <= 9; i++ {
		var addr types.Address
		addr[19] = byte(i)
		if !a.IsWarmAddress(addr) {
			t.Errorf("precompile 0x%02x should be warm at start", i)
		}
	}

	// An arbitrary non-precompile must NOT be warm.
	var cold types.Address
	cold[19] = 0x0A
	if a.IsWarmAddress(cold) {
		t.Error("address 0x0A should not be warm at start")
	}
}

// TestPrecompiledAddressesSet checks the package-level PrecompiledAddresses
// map has exactly the 9 expected entries.
func TestPrecompiledAddressesSet(t *testing.T) {
	if len(PrecompiledAddresses) != 9 {
		t.Errorf("expected 9 precompiled addresses, got %d", len(PrecompiledAddresses))
	}
	for i := 1; i <= 9; i++ {
		var addr types.Address
		addr[19] = byte(i)
		if !PrecompiledAddresses[addr] {
			t.Errorf("missing precompile address 0x%02x", i)
		}
	}
}

// TestWarmUpAddress verifies that WarmUpAddress marks an address warm and
// that IsWarmAddress reflects it.
func TestWarmUpAddress(t *testing.T) {
	a := NewAccruedSubstate()
	addr := types.Address{0x42}

	if a.IsWarmAddress(addr) {
		t.Error("address should be cold before WarmUpAddress")
	}
	a.WarmUpAddress(addr)
	if !a.IsWarmAddress(addr) {
		t.Error("address should be warm after WarmUpAddress")
	}
}

// TestWarmUpStorage verifies that WarmUpStorage marks a (address, slot) pair warm.
func TestWarmUpStorage(t *testing.T) {
	a := NewAccruedSubstate()
	addr := types.Address{0x01}
	slot := types.Hash{0xFF}

	if a.IsWarmStorage(addr, slot) {
		t.Error("slot should be cold before WarmUpStorage")
	}
	a.WarmUpStorage(addr, slot)
	if !a.IsWarmStorage(addr, slot) {
		t.Error("slot should be warm after WarmUpStorage")
	}
}

// TestWarmStorageKeyUniqueness verifies that (addr1, slot) and (addr2, slot)
// are tracked independently.
func TestWarmStorageKeyUniqueness(t *testing.T) {
	a := NewAccruedSubstate()
	addr1 := types.Address{0x01}
	addr2 := types.Address{0x02}
	slot := types.Hash{0xAB}

	a.WarmUpStorage(addr1, slot)

	if !a.IsWarmStorage(addr1, slot) {
		t.Error("addr1/slot should be warm")
	}
	if a.IsWarmStorage(addr2, slot) {
		t.Error("addr2/slot should still be cold")
	}
}

// TestMergeSelfDestructs verifies that child SelfDestructs are unioned into parent.
func TestMergeSelfDestructs(t *testing.T) {
	parent := NewAccruedSubstate()
	child := NewAccruedSubstate()

	addr := types.Address{0xAA}
	child.SelfDestructs[addr] = true

	parent.Merge(child)

	if !parent.SelfDestructs[addr] {
		t.Error("parent should contain child's self-destruct after Merge")
	}
}

// TestMergeLogs verifies that child logs are appended in order.
func TestMergeLogs(t *testing.T) {
	parent := NewAccruedSubstate()
	child := NewAccruedSubstate()

	logA := Log{Address: types.Address{0x01}, Topics: []types.Hash{{0x01}}, Data: []byte{1}}
	logB := Log{Address: types.Address{0x02}, Topics: []types.Hash{{0x02}}, Data: []byte{2}}
	parent.Logs = append(parent.Logs, logA)
	child.Logs = append(child.Logs, logB)

	parent.Merge(child)

	if len(parent.Logs) != 2 {
		t.Fatalf("expected 2 logs after merge, got %d", len(parent.Logs))
	}
	if parent.Logs[0].Address != logA.Address || parent.Logs[1].Address != logB.Address {
		t.Error("logs not in expected order after merge")
	}
}

// TestMergeTouchedAccounts verifies that touched accounts are unioned.
func TestMergeTouchedAccounts(t *testing.T) {
	parent := NewAccruedSubstate()
	child := NewAccruedSubstate()

	addrP := types.Address{0x01}
	addrC := types.Address{0x02}
	parent.TouchedAccounts[addrP] = true
	child.TouchedAccounts[addrC] = true

	parent.Merge(child)

	if !parent.TouchedAccounts[addrP] || !parent.TouchedAccounts[addrC] {
		t.Error("touched accounts not fully unioned after merge")
	}
}

// TestMergeRefund verifies that child refunds are summed into the parent.
func TestMergeRefund(t *testing.T) {
	parent := NewAccruedSubstate()
	child := NewAccruedSubstate()

	parent.Refund = 100
	child.Refund = 50

	parent.Merge(child)

	if parent.Refund != 150 {
		t.Errorf("expected refund 150 after merge, got %d", parent.Refund)
	}
}

// TestMergeWarmAddresses verifies that child warm addresses are unioned into parent.
func TestMergeWarmAddresses(t *testing.T) {
	parent := NewAccruedSubstate()
	child := NewAccruedSubstate()

	extra := types.Address{0xFF}
	child.WarmUpAddress(extra)

	parent.Merge(child)

	if !parent.IsWarmAddress(extra) {
		t.Error("parent should inherit child's warm address after merge")
	}
}

// TestMergeWarmStorage verifies that child warm storage slots are unioned into parent.
func TestMergeWarmStorage(t *testing.T) {
	parent := NewAccruedSubstate()
	child := NewAccruedSubstate()

	addr := types.Address{0xBB}
	slot := types.Hash{0xCC}
	child.WarmUpStorage(addr, slot)

	parent.Merge(child)

	if !parent.IsWarmStorage(addr, slot) {
		t.Error("parent should inherit child's warm storage slot after merge")
	}
}

// TestMergeIdempotentDuplicates verifies that merging overlapping sets does
// not produce duplicate entries or panic.
func TestMergeIdempotentDuplicates(t *testing.T) {
	parent := NewAccruedSubstate()
	child := NewAccruedSubstate()

	addr := types.Address{0x01}
	slot := types.Hash{0x01}
	parent.WarmUpAddress(addr)
	child.WarmUpAddress(addr)
	parent.WarmUpStorage(addr, slot)
	child.WarmUpStorage(addr, slot)

	parent.Merge(child)

	// Sets are deduplicated by map semantics — just check no panic and correct state.
	if !parent.IsWarmAddress(addr) {
		t.Error("address should remain warm after duplicate merge")
	}
	if !parent.IsWarmStorage(addr, slot) {
		t.Error("storage slot should remain warm after duplicate merge")
	}
}
