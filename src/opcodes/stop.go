package opcodes

import "github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"

func OpSTOP(e types.Executor) (types.OpResult, error) {
	return types.OpResult{Halt: types.HaltStop}, nil
}