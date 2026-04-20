package opcodes

import (
	"math/big"

	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

func OpPOP(e types.Executor) error {
	_, err := e.GetStack().Pop()
	if err != nil {
		return err
	}
	e.SetPC(e.GetPC() + 1)
	return nil
}

// MakePush returns an OpFunc that pushes n bytes from bytecode onto the stack.
func MakePush(n int) types.OpFunc {
	return func(e types.Executor) error {
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
		return e.GetStack().Push(new(big.Int).SetBytes(val))
	}
}
