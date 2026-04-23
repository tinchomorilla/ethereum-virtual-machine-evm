package stack

import (
	"errors"
	"math/big"
)

const (
	// MaxStackDepth defines the strict limit for the EVM stack.
	MaxStackDepth = 1024
)

var (
	ErrStackOverflow  = errors.New("stack overflow")
	ErrStackUnderflow = errors.New("stack underflow")
)

// EVMStack implements the stack data structure for the Ethereum Virtual Machine.
type EVMStack struct {
	data []*big.Int
}

// New creates a new EVM stack with pre-allocated memory to avoid reallocation overhead.
func New() *EVMStack {
	return &EVMStack{
		// Pre-allocate the slice capacity to 1024 to save Garbage Collector cycles.
		data: make([]*big.Int, 0, MaxStackDepth),
	}
}

// Push adds a new 256-bit word to the top of the stack.
func (st *EVMStack) Push(d *big.Int) error {
	if len(st.data) >= MaxStackDepth {
		return ErrStackOverflow
	}
	st.data = append(st.data, d)
	return nil
}

// Pop removes and returns the 256-bit word at the top of the stack.
func (st *EVMStack) Pop() (*big.Int, error) {
	l := len(st.data)
	if l == 0 {
		return nil, ErrStackUnderflow
	}
	// Extract the top element
	res := st.data[l-1]
	// Remove the top element from the slice
	st.data = st.data[:l-1]
	return res, nil
}

// Peek returns the n-th item from the top of the stack without removing it.
// n=1 returns the top item, n=2 returns the second to top, etc.
func (st *EVMStack) Peek(n int) (*big.Int, error) {
	l := len(st.data)
	if n <= 0 || n > l {
		return nil, ErrStackUnderflow
	}
	return st.data[l-n], nil
}

// Swap swaps the top element with the (n+1)-th element from the top.
// Swap(1) exchanges positions 1 and 2, Swap(2) exchanges 1 and 3, etc.
func (st *EVMStack) Swap(n int) error {
	l := len(st.data)
	if n+1 > l {
		return ErrStackUnderflow
	}
	st.data[l-1], st.data[l-1-n] = st.data[l-1-n], st.data[l-1]
	return nil
}

// Len returns the current number of items in the stack.
func (st *EVMStack) Len() int {
	return len(st.data)
}