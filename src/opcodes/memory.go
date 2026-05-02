package opcodes

import (
	"math/big"

	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

// padTo32Bytes converts a *big.Int to a 32-byte array, padding with leading zeros.
// This is required because MSTORE always writes exactly 32 bytes (1 word) to memory,
// regardless of the actual mathematical size of the number.
func padTo32Bytes(val *big.Int) []byte {
	buf := make([]byte, 32)
	return val.FillBytes(buf)
}

func OpMSTORE(e types.Executor) (types.OpResult, error) {
	stack := e.GetStack()
	memory := e.GetMemory()

	// MSTORE takes two arguments: offset and value
	offset, err := stack.Pop()
	if err != nil {
		return types.OpResult{}, err
	}
	value, err := stack.Pop()
	if err != nil {
		return types.OpResult{}, err
	}

	// Pad the value to 32 bytes (1 word)
	paddedValue := padTo32Bytes(value)

	newSize := offset.Uint64() + 32
	if newSize > memory.Len() {
		words := (newSize + 31) / 32
		memory.Resize(words * 32)
	}

	// Write the padded value to memory at the specified offset
	err = memory.Set(offset.Uint64(), 32, paddedValue)
	if err != nil {
		return types.OpResult{}, err
	}

	e.SetPC(e.GetPC() + 1)
	return types.OpResult{}, nil

}

func OpMLOAD(e types.Executor) (types.OpResult, error) {
	stack := e.GetStack()
	memory := e.GetMemory()

	// MLOAD takes one argument: offset
	offset, err := stack.Pop()
	if err != nil {
		return types.OpResult{}, err
	}

	newSize := offset.Uint64() + 32
	if newSize > memory.Len() {
		words := (newSize + 31) / 32
		memory.Resize(words * 32)
	}

	// Read 32 bytes from memory at the specified offset
	data, err := memory.Get(offset.Uint64(), 32)
	if err != nil {
		return types.OpResult{}, err
	}

	// Push the loaded value onto the stack as a big.Int
	err = stack.Push(new(big.Int).SetBytes(data))
	if err != nil {
		return types.OpResult{}, err
	}
	e.SetPC(e.GetPC() + 1)
	return types.OpResult{}, nil
}

func OpMSTORE8(e types.Executor) (types.OpResult, error) {
	stack := e.GetStack()
	memory := e.GetMemory()

	// MSTORE8 takes two arguments: offset and value
	offset, err := stack.Pop()
	if err != nil {
		return types.OpResult{}, err
	}
	value, err := stack.Pop()
	if err != nil {
		return types.OpResult{}, err
	}

	// Write the least significant byte of the value to memory at the specified offset
	byteValue := byte(value.Uint64() & 0xff)

	newSize := offset.Uint64() + 1
	if newSize > memory.Len() {
		words := (newSize + 31) / 32
		memory.Resize(words * 32)
	}

	err = memory.Set(offset.Uint64(), 1, []byte{byteValue})
	if err != nil {
		return types.OpResult{}, err
	}

	e.SetPC(e.GetPC() + 1)
	return types.OpResult{}, nil
}

func OpMSIZE(e types.Executor) (types.OpResult, error) {
	stack := e.GetStack()
	memory := e.GetMemory()

	// MSIZE pushes the current size of memory (in bytes) onto the stack
	err := stack.Push(new(big.Int).SetUint64(memory.Len()))
	if err != nil {
		return types.OpResult{}, err
	}
	e.SetPC(e.GetPC() + 1)
	return types.OpResult{}, nil
}
