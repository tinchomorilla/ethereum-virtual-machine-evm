package substate

import (
	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

// TODO: In the future, I need to implement variable inheritance
// between the parent and child. Right now, the child is borned with empty variables,
// which is incorrect.

// StorageKey identifies a single storage slot: a (address, slot) pair.
// Used as the key type for the warm-storage set AK.
type StorageKey struct {
	Address types.Address
	Slot    types.Hash
}

// Log represents a single log entry emitted by the LOG opcodes.
type Log struct {
	Address types.Address // contract that emitted the log
	Topics  []types.Hash  // indexed topics (0..4)
	Data    []byte        // unindexed data
}

// AccruedSubstate maps exactly to the tuple A ≡ (As, Al, At, Ar, Aa, AK) (defined in Yellow Paper)
// It accumulates side-effects that must be applied or discarded atomically
// when a call frame succeeds or reverts.
type AccruedSubstate struct {
	// As — set of accounts scheduled for self-destruction at the end of the transaction.
	SelfDestructs map[types.Address]bool

	// Al — ordered series of log entries created during execution.
	Logs []Log

	// At — set of touched accounts (used for empty-account deletion, EIP-161).
	TouchedAccounts map[types.Address]bool

	// Ar — accumulated refund counter (gas refunds from SSTORE clears, SELFDESTRUCT).
	Refund uint64

	// Aa — set of warm addresses per EIP-2929 (accessed address set).
	WarmAddresses map[types.Address]bool

	// AK — set of warm storage slots per EIP-2929 (accessed storage keys).
	WarmStorage map[StorageKey]bool
}

// PrecompiledAddresses is the set π of precompiled contract addresses (0x01..0x09).
// These are always considered warm from the start of every transaction per EIP-2929.
var PrecompiledAddresses = func() map[types.Address]bool {
	m := make(map[types.Address]bool)
	for i := 1; i <= 9; i++ {
		var addr types.Address
		addr[19] = byte(i)
		m[addr] = true
	}
	return m
}()

// NewAccruedSubstate returns the empty substate A⁰:
//
//	A⁰ ≡ (∅, (), ∅, 0, π, ∅)
//
// All maps are pre-allocated and WarmAddresses is seeded with the precompiled addresses π.
func NewAccruedSubstate() *AccruedSubstate {
	warm := make(map[types.Address]bool)
	for addr := range PrecompiledAddresses {
		warm[addr] = true
	}
	return &AccruedSubstate{
		SelfDestructs:   make(map[types.Address]bool),
		Logs:            []Log{},
		TouchedAccounts: make(map[types.Address]bool),
		Refund:          0,
		WarmAddresses:   warm,
		WarmStorage:     make(map[StorageKey]bool),
	}
}

// Merge incorporates a successful child call frame's substate into the parent,
// Set fields are unioned, logs are appended in order, and refunds are summed.
func (a *AccruedSubstate) Merge(child *AccruedSubstate) {
	for addr := range child.SelfDestructs {
		a.SelfDestructs[addr] = true
	}
	a.Logs = append(a.Logs, child.Logs...)
	for addr := range child.TouchedAccounts {
		a.TouchedAccounts[addr] = true
	}
	a.Refund += child.Refund
	for addr := range child.WarmAddresses {
		a.WarmAddresses[addr] = true
	}
	for key := range child.WarmStorage {
		a.WarmStorage[key] = true
	}
}

// IsWarmAddress reports whether addr is in the warm address set Aa (EIP-2929).
func (a *AccruedSubstate) IsWarmAddress(addr types.Address) bool {
	return a.WarmAddresses[addr]
}

// IsWarmStorage reports whether the (addr, slot) pair is in the warm storage set AK (EIP-2929).
func (a *AccruedSubstate) IsWarmStorage(addr types.Address, slot types.Hash) bool {
	return a.WarmStorage[StorageKey{addr, slot}]
}

// WarmUpAddress adds addr to the warm address set Aa.
func (a *AccruedSubstate) WarmUpAddress(addr types.Address) {
	a.WarmAddresses[addr] = true
}

// WarmUpStorage adds the (addr, slot) pair to the warm storage set AK.
func (a *AccruedSubstate) WarmUpStorage(addr types.Address, slot types.Hash) {
	a.WarmStorage[StorageKey{addr, slot}] = true
}

// AddRefund adds the given amount to the accumulated refund counter Ar.
func (a *AccruedSubstate) AddRefund(amount uint64) {
	a.Refund += amount
}

// SubRefund subtracts the given amount from the accumulated refund counter Ar.
func (a *AccruedSubstate) SubRefund(amount uint64) {
	a.Refund -= amount
}
