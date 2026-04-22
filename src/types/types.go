package types

import (
	"errors"
	"math/big"
)

// ErrStopExecution signals a normal STOP halt (not a real error).
var ErrStopExecution = errors.New("stop execution")

// OpFunc is a function that implements a single EVM opcode.
type OpFunc func(Executor) error

// Executor is the interface opcode implementations use to interact with the EVM.
// It decouples the opcodes package from the core package, avoiding circular imports.
type Executor interface {
	GetStack() Stack
	GetMemory() Memory
	GetCode() []byte
	GetPC() uint64
	SetPC(uint64)
	GetJumpDests() map[uint64]struct{}
}

// OpCode represents an EVM instruction.
type OpCode byte

// Address represents the 20-byte address of an Ethereum account.
type Address [20]byte

// Hash represents the 32-byte Keccak-256 hash.
type Hash [32]byte

// Stack defines the interface for the EVM execution stack.
// The EVM stack can hold up to 1024 items, each being a 256-bit word.
type Stack interface {
	Push(d *big.Int) error
	Pop() (*big.Int, error)
	Peek(n int) (*big.Int, error)
	Len() int
}

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
}

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

	GetState(addr Address, key Hash) Hash
	SetState(addr Address, key Hash, value Hash)
}