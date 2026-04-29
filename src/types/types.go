package types

import (
	"errors"
	"math/big"
)

// ErrStopExecution signals a normal STOP halt (not a real error).
var ErrStopExecution = errors.New("stop execution")

// OpFunc is a function that implements a single EVM opcode.
type OpFunc func(Executor) error

// OpCode represents an EVM instruction.
type OpCode byte

// Address represents the 20-byte address of an Ethereum account.
type Address [20]byte

// Hash represents the 32-byte Keccak-256 hash.
type Hash [32]byte


// MachineState holds the volatile, execution-specific state of the EVM.
// It is completely wiped after the execution halts.
type MachineState struct {
	// Pc is the Program Counter, indicating the current instruction index.
	Pc uint64
	
	// Gas is the available gas for the current execution context.
	Gas uint64
	
	// Stack is the 1024-item execution stack.
	Stack Stack
	
	// Memory is the volatile, expandable byte array.
	Memory Memory
	
	// ReturnData holds the output of the previous sub-context call.
	ReturnData []byte
}

// BlockContext holds the block-level information that is constant for every call in a block.
type BlockContext struct {
	// Hc: Address of the block beneficiary (miner/validator).
	Coinbase Address

	// Hs: Unix timestamp of the current block.
	Timestamp uint64

	// Hi: Current block number.
	Number uint64

	// Hp: Previous RANDAO mix (post-merge) or difficulty (pre-merge).
	PrevRandao Hash

	// Hl: Gas limit of the current block.
	GasLimit uint64

	// ChainID as defined by EIP-155.
	ChainID *big.Int

	// BaseFee per gas unit for the current block (EIP-1559).
	BaseFee *big.Int

	// GetHash returns the hash of block n (only valid for the last 256 blocks).
	GetHash func(uint64) Hash
}

// ExecutionContext represents the execution environment (tuple I) defined in the Yellow Paper.
// It contains information that remains constant during the execution of a specific context.
type ExecutionContext struct {
	// Ia: The address of the account which owns the code that is executing.
	Address Address
	
	// Io: The sender address of the transaction that originated this execution.
	Origin Address

	// Ip: The price of gas paid by the signer of the transaction.
	GasPrice *big.Int
	
	// Id: The byte array that is the input data to this execution (calldata).
	Input []byte
	
	// Is: The address of the account which caused the code to be executing (caller).
	Caller Address
	
	// Iv: The value, in Wei, passed to this account as part of the execution.
	Value *big.Int
	
	// Ib: The byte array that is the machine code to be executed.
	ByteCode []byte
	
	// Ie: The depth of the present message-call or contract-creation.
	Depth int
	
	// Iw: The permission to make modifications to the state (false for STATICCALL).
	ReadOnly bool

	// StateDB provides access to the global world state.
	StateDB StateDB

	// Block holds the block-level context for this execution.
	Block BlockContext
}

//------------------------ Interfaces -----------------------------------------//

// Memory defines the interface for the EVM volatile memory.
// Memory is a byte array that expands as needed, incurring a gas cost.
type Memory interface {
	Set(offset uint64, size uint64, value []byte) error
	Get(offset uint64, size uint64) ([]byte, error)
	Len() uint64
	Resize(size uint64)
}

// StateDB defines the interface for the EVM world state access.
// It acts as an abstraction layer over the Merkle Patricia Trie.
type StateDB interface {
	GetBalance(addr Address) *big.Int
	AddBalance(addr Address, amount *big.Int)
	SubBalance(addr Address, amount *big.Int)
	GetCodeSize(addr Address) uint64
	GetCodeHash(addr Address) Hash
	GetState(addr Address, key Hash) Hash
	SetState(addr Address, key Hash, value Hash)
	GetCommittedState(addr Address, key Hash) Hash
}

// Stack defines the interface for the EVM execution stack.
// The EVM stack can hold up to 1024 items, each being a 256-bit word.
type Stack interface {
	Push(d *big.Int) error
	Pop() (*big.Int, error)
	Peek(n int) (*big.Int, error)
	Swap(n int) error
	Len() int
}

// Executor is the interface opcode implementations use to interact with the EVM.
// It decouples the opcodes package from the core package, avoiding circular imports.
type Executor interface {
	GetStack() Stack
	GetMemory() Memory
	GetCode() []byte
	GetPC() uint64
	SetPC(uint64)
	GetJumpDests() map[uint64]struct{}
	GetContext() ExecutionContext
	GetGas() uint64
	GetReturnData() []byte
}