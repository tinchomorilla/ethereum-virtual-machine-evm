package gas

import (
	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

// Gas constants for EIP-2929 and EIP-2200
const (
	GZero     uint64 = 0
	GJumpDest uint64 = 1
	GBase     uint64 = 2
	GVeryLow  uint64 = 3
	GLow      uint64 = 5
	GMid      uint64 = 8
	GHigh     uint64 = 10
	GExp      uint64 = 10
	GExpByte  uint64 = 10
	GSha3     uint64 = 30
	GSha3Word uint64 = 6
	GBalance  uint64 = 20
	GLog      uint64 = 375
	GLogTopic uint64 = 375
	GLogData  uint64 = 8

	// EIP-2929 constants
	GColdSload       uint64 = 2100
	GWarmStorageRead uint64 = 100

	// EIP-2200 constants
	GSset   uint64 = 20000
	GSreset uint64 = 2900
	RSClear uint64 = 15000

	GSelfDestruct uint64 = 5000
	GCreate       uint64 = 32000
	GCodeDeposit  uint64 = 200
	GCall         uint64 = 40
	GCallValue    uint64 = 9000
	GCallStipend  uint64 = 2300
	GNewAccount   uint64 = 25000
	GCopy         uint64 = 3
	GMemory       uint64 = 3
	GBlockHash    uint64 = 20
	GExtCode      uint64 = 20
)

// dynamicCostFunc defines a function that calculates the runtime gas cost of an opcode.
// It returns the additional dynamic cost, or an error if stack requirements aren't met.
type dynamicCostFunc func(e types.Executor) (uint64, error)

// dynamicCost holds the dynamic gas cost functions for opcodes that require runtime calculation.
var dynamicCost [256]dynamicCostFunc

// staticCost holds the base gas cost for each opcode (0x00..0xff).
// Opcodes with dynamic behavior store only their static component here.
var staticCost [256]uint64

func init() {

	staticCost[0x00] = GZero // STOP
	// Aritmetic
	staticCost[0x01] = GVeryLow // ADD
	staticCost[0x02] = GLow     // MUL
	staticCost[0x03] = GVeryLow // SUB
	staticCost[0x04] = GLow     // DIV
	staticCost[0x05] = GLow     // SDIV
	staticCost[0x06] = GLow     // MOD
	staticCost[0x07] = GLow     // SMOD
	staticCost[0x08] = GMid     // ADDMOD
	staticCost[0x09] = GMid     // MULMOD
	staticCost[0x0a] = GExp     // EXP
	staticCost[0x0b] = GLow     // SIGNEXTEND

	// 0x10: comparison and bitwise
	for op := byte(0x10); op <= 0x1d; op++ {
		staticCost[op] = GVeryLow
	}

	// 0x20: KECCAK256
	staticCost[0x20] = GSha3 // SHA3

	// 0x30: environmental information
	staticCost[0x30] = GBase    // ADDRESS
	staticCost[0x31] = GBalance // BALANCE
	staticCost[0x32] = GBase    // ORIGIN
	staticCost[0x33] = GBase    // CALLER
	staticCost[0x34] = GBase    // CALLVALUE
	staticCost[0x35] = GVeryLow // CALLDATALOAD
	staticCost[0x36] = GBase    // CALLDATASIZE
	staticCost[0x37] = GVeryLow // CALLDATACOPY
	staticCost[0x38] = GBase    // CODESIZE
	staticCost[0x39] = GVeryLow // CODECOPY
	staticCost[0x3a] = GBase    // GASPRICE
	staticCost[0x3b] = GExtCode // EXTCODESIZE
	staticCost[0x3c] = GExtCode // EXTCODECOPY
	staticCost[0x3d] = GBase    // RETURNDATASIZE
	staticCost[0x3e] = GVeryLow // RETURNDATACOPY
	staticCost[0x3f] = GExtCode // EXTCODEHASH

	// 0x40: block information
	staticCost[0x40] = GBlockHash // BLOCKHASH
	staticCost[0x41] = GBase      // COINBASE
	staticCost[0x42] = GBase      // TIMESTAMP
	staticCost[0x43] = GBase      // NUMBER
	staticCost[0x44] = GBase      // PREVRANDAO
	staticCost[0x45] = GBase      // GASLIMIT
	staticCost[0x46] = GBase      // CHAINID
	staticCost[0x47] = GLow       // SELFBALANCE
	staticCost[0x48] = GBase      // BASEFEE
	staticCost[0x49] = GBase      // BLOBHASH
	staticCost[0x4a] = GBase      // BLOBBASEFEE

	// 0x50: stack, memory, storage and flow
	staticCost[0x50] = GBase     // POP
	staticCost[0x51] = GVeryLow  // MLOAD
	staticCost[0x52] = GVeryLow  // MSTORE
	staticCost[0x53] = GVeryLow  // MSTORE8
	staticCost[0x56] = GMid      // JUMP
	staticCost[0x57] = GHigh     // JUMPI
	staticCost[0x58] = GBase     // PC
	staticCost[0x59] = GBase     // MSIZE
	staticCost[0x5a] = GBase     // GAS
	staticCost[0x5b] = GJumpDest // JUMPDEST

	// 0x60: PUSH1..PUSH32
	for op := byte(0x60); op <= 0x7f; op++ {
		staticCost[op] = GVeryLow
	}

	// 0x80: DUP1..DUP16
	for op := byte(0x80); op <= 0x8f; op++ {
		staticCost[op] = GVeryLow
	}

	// 0x90: SWAP1..SWAP16
	for op := byte(0x90); op <= 0x9f; op++ {
		staticCost[op] = GVeryLow
	}

	// 0xa0: LOG0..LOG4
	for op := byte(0xa0); op <= 0xa4; op++ {
		staticCost[op] = GLog + uint64(op-0xa0)*GLogTopic
	}

	// 0xf0: create/call family and halting
	staticCost[0xf0] = GCreate       // CREATE
	staticCost[0xf1] = GCall         // CALL
	staticCost[0xf2] = GCall         // CALLCODE
	staticCost[0xf3] = GZero         // RETURN
	staticCost[0xf4] = GCall         // DELEGATECALL
	staticCost[0xf5] = GCreate       // CREATE2
	staticCost[0xfa] = GCall         // STATICCALL
	staticCost[0xfd] = GZero         // REVERT
	staticCost[0xff] = GSelfDestruct // SELFDESTRUCT

	// Dynamic cost functions
	dynamicCost[0x51] = gasMStoreAndMLoad
	dynamicCost[0x52] = gasMStoreAndMLoad
	dynamicCost[0x53] = gasMStore8

	// Dynamic costs for storage (EIP-2929 / EIP-2200)
	dynamicCost[0x54] = gasSLoad
	dynamicCost[0x55] = gasSStore

	// Dynamic costs for RETURN and REVERT
	dynamicCost[0xf3] = gasReturnAndRevert // RETURN
	dynamicCost[0xfd] = gasReturnAndRevert // REVERT
}

// gasSLoad calculates gas for SLOAD (0x54)
func gasSLoad(e types.Executor) (uint64, error) {
	keyInt, err := e.GetStack().Peek(1)
	if err != nil {
		return 0, err
	}
	key := types.BigIntToHash(keyInt)
	addr := e.GetContext().Address

	if e.GetAccruedSubstate().IsWarmStorage(addr, key) {
		return GWarmStorageRead, nil
	}
	return GColdSload, nil
}

// gasSStore calculates gas for SSTORE (0x55).
// v0 = original value (before transaction)
// v = current value (before SSTORE)
// v' = new value (after SSTORE)
func gasSStore(e types.Executor) (uint64, error) {
	keyInt, err := e.GetStack().Peek(1)
	if err != nil {
		return 0, err
	}
	valInt, err := e.GetStack().Peek(2)
	if err != nil {
		return 0, err
	}

	key := types.BigIntToHash(keyInt)
	vPrime := types.BigIntToHash(valInt) // New value
	ctx := e.GetContext()
	addr := ctx.Address
	db := ctx.StateDB

	substate := e.GetAccruedSubstate()
	isWarm := substate.IsWarmStorage(addr, key)

	var v0, v types.Hash
	if db != nil {
		v0 = db.GetCommittedState(addr, key) // Original value
		v = db.GetState(addr, key)           // Current value
	}

	var cost uint64

	// 1. Access cost
	if !isWarm {
		cost += GColdSload
	}

	empty := types.Hash{} // Helper for the 32-byte zero value

	// 2. Writing logic and Refunds
	if v == vPrime {
		// No change to current value
		cost += GWarmStorageRead
	} else {
		if v0 == v {
			// First time changing the slot in this transaction
			if v0 == empty {
				cost += GSset // Creating a new slot
			} else {
				cost += GSreset // Updating an existing slot
			}
		} else {
			// Subsequent changes in the same transaction
			cost += GWarmStorageRead
		}

		// --- Refund Calculation --- //

		// Rule 1: The Clear (v != 0, vPrime == 0)
		if v != empty && vPrime == empty {
			substate.AddRefund(RSClear)
		}

		// Rule 2: The Regret (v0 != 0, v == 0, vPrime != 0)
		if v0 != empty && v == empty && vPrime != empty {
			substate.SubRefund(RSClear)
		}

		// Rule 3: The Restore (v0 == vPrime, but v was different)
		if v0 == vPrime && v0 != v {
			if v0 == empty {
				substate.AddRefund(GSset - GWarmStorageRead)
			} else {
				substate.AddRefund(GSreset - GWarmStorageRead)
			}
		}
	}

	return cost, nil
}

// memoryCost calculates the gas cost for a given memory size in words (32 bytes).
func memoryCost(words uint64) uint64 {
	return (words * GMemory) + ((words * words) / 512)
}

// calcMemExpansionCost calculates the additional gas cost for expanding memory from currentSize to newSize (in bytes).
func calcMemExpansionCost(currentSize uint64, newSize uint64) uint64 {
	if newSize <= currentSize {
		return 0
	}

	currentWords := (currentSize + 31) / 32
	newWords := (newSize + 31) / 32

	// The cost is the difference in memory cost before and after expansion.
	return memoryCost(newWords) - memoryCost(currentWords)
}

func gasMStoreAndMLoad(e types.Executor) (uint64, error) {
	offset, err := e.GetStack().Peek(1)
	if err != nil {
		return 0, err
	}

	size := uint64(32)

	// i should check if the offset is valid (not too large) but for simplicity, let's assume it's always valid.
	newSize := offset.Uint64() + size

	currentSize := e.GetMemory().Len()

	return calcMemExpansionCost(currentSize, newSize), nil
}

// gasReturnAndRevert calculates the dynamic gas cost for RETURN (0xf3) and REVERT (0xfd).
// It peeks at the stack to determine memory expansion but does NOT pop elements.
func gasReturnAndRevert(e types.Executor) (uint64, error) {
	// RETURN/REVERT expects: offset, size
	offset, err := e.GetStack().Peek(1)
	if err != nil {
		return 0, err
	}
	size, err := e.GetStack().Peek(2)
	if err != nil {
		return 0, err
	}

	// If size is 0, there is no memory expansion, regardless of the offset.
	if size.Sign() == 0 {
		return 0, nil
	}

	newSize := offset.Uint64() + size.Uint64()
	currentSize := e.GetMemory().Len()

	return calcMemExpansionCost(currentSize, newSize), nil
}

func gasMStore8(e types.Executor) (uint64, error) {
	offset, err := e.GetStack().Peek(1)
	if err != nil {
		return 0, err
	}

	size := uint64(1)

	newSize := offset.Uint64() + size
	currentSize := e.GetMemory().Len()

	return calcMemExpansionCost(currentSize, newSize), nil
}
