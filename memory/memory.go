package memory

import "errors"

var (
	ErrMemoryOutOfBounds = errors.New("memory access out of bounds")
)

// EVMMemory implements the volatile memory for the EVM.
// It is a simple byte array that starts empty and grows as needed.
type EVMMemory struct {
	data []byte
}

// New creates a new empty memory instance.
func New() *EVMMemory {
	return &EVMMemory{
		data: make([]byte, 0),
	}
}

// Resize expands the memory to the specified target size.
// In the EVM, memory always expands in multiples of 32 bytes (1 word).
// The caller (core EVM loop) is responsible for calculating the correct word-aligned size.
func (m *EVMMemory) Resize(targetSize uint64) {
	if uint64(len(m.data)) < targetSize {
		// Append necessary zeros to reach the target size
		m.data = append(m.data, make([]byte, targetSize-uint64(len(m.data)))...)
	}
}

// Set writes a slice of bytes into the memory at the given offset.
// It assumes the memory has already been resized and paid for in gas.
func (m *EVMMemory) Set(offset uint64, size uint64, value []byte) error {
	if size == 0 {
		return nil
	}
	if offset+size > uint64(len(m.data)) {
		return ErrMemoryOutOfBounds
	}

	// Copy the value into memory. If value is smaller than size,
	// the remaining bytes are left untouched (already zeroed by Resize).
	copy(m.data[offset:offset+size], value)
	return nil
}

// Get returns a copy of a memory segment.
func (m *EVMMemory) Get(offset uint64, size uint64) ([]byte, error) {
	if size == 0 {
		return []byte{}, nil
	}
	if offset+size > uint64(len(m.data)) {
		return nil, ErrMemoryOutOfBounds
	}

	// Return a copy to prevent external mutation of the internal state
	cpy := make([]byte, size)
	copy(cpy, m.data[offset:offset+size])
	return cpy, nil
}

// Len returns the current active size of the memory.
func (m *EVMMemory) Len() uint64 {
	return uint64(len(m.data))
}