package types

import "math/big"

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