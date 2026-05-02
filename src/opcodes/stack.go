package opcodes

import (
	"math/big"

	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

// MakeDup returns an OpFunc for DUP_n: duplicates the n-th item from the top.
// DUP1 (n=1) duplicates the top, DUP16 (n=16) duplicates the 16th item.
func MakeDup(n int) types.OpFunc {
	return func(e types.Executor) (types.OpResult, error) {
		val, err := e.GetStack().Peek(n)
		if err != nil {
			return types.OpResult{}, err
		}
		if err := e.GetStack().Push(new(big.Int).Set(val)); err != nil {
			return types.OpResult{}, err
		}
		e.SetPC(e.GetPC() + 1)
		return types.OpResult{}, nil
	}
}

// MakeSwap returns an OpFunc for SWAP_n: swaps the top with the (n+1)-th item.
// SWAP1 (n=1) swaps positions 1 and 2, SWAP16 (n=16) swaps positions 1 and 17.
func MakeSwap(n int) types.OpFunc {
	return func(e types.Executor) (types.OpResult, error) {
		if err := e.GetStack().Swap(n); err != nil {
			return types.OpResult{}, err
		}
		e.SetPC(e.GetPC() + 1)
		return types.OpResult{}, nil
	}
}

func OpPOP(e types.Executor) (types.OpResult, error) {
	_, err := e.GetStack().Pop()
	if err != nil {
		return types.OpResult{}, err
	}
	e.SetPC(e.GetPC() + 1)
	return types.OpResult{}, nil
}

// MakePush returns an OpFunc that pushes n bytes from bytecode onto the stack.
func MakePush(n int) types.OpFunc {
	return func(e types.Executor) (types.OpResult, error) {
		code := e.GetCode()
		start := e.GetPC() + 1
		end := start + uint64(n)

		var val []byte
		if end <= uint64(len(code)) {
			val = code[start:end]
		} else {
			// Pad with leading zeros if bytecode is shorter than expected.
			val = make([]byte, n)
			if start < uint64(len(code)) {
				copy(val[uint64(n)-(uint64(len(code))-start):], code[start:])
			}
		}

		e.SetPC(e.GetPC() + uint64(n) + 1)
		if err := e.GetStack().Push(new(big.Int).SetBytes(val)); err != nil {
			return types.OpResult{}, err
		}
		return types.OpResult{}, nil
	}
}
