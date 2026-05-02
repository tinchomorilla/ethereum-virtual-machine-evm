package opcodes

import "github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"

func OpSTOP(e types.Executor) (types.OpResult, error) {
	return types.OpResult{Halt: types.HaltStop}, nil
}

func OpRETURN(e types.Executor) (types.OpResult, error) {
	stack := e.GetStack()
	offset, size, err := popTwoArgs(stack)
	if err != nil {
		return types.OpResult{}, err
	}

	newSize := offset.Uint64() + size.Uint64()
	if newSize > e.GetMemory().Len() {
		words := (newSize + 31) / 32
		e.GetMemory().Resize(words * 32)
	}

	data, err := e.GetMemory().Get(offset.Uint64(), size.Uint64())
	if err != nil {
		return types.OpResult{}, err
	}
	e.SetReturnData(data)
	return types.OpResult{Halt: types.HaltReturn}, nil
}

func OpREVERT(e types.Executor) (types.OpResult, error) {
	stack := e.GetStack()
	offset, size, err := popTwoArgs(stack)
	if err != nil {
		return types.OpResult{}, err
	}

	newSize := offset.Uint64() + size.Uint64()
	if newSize > e.GetMemory().Len() {
		words := (newSize + 31) / 32
		e.GetMemory().Resize(words * 32)
	}

	data, err := e.GetMemory().Get(offset.Uint64(), size.Uint64())
	if err != nil {
		return types.OpResult{}, err
	}
	e.SetReturnData(data)
	return types.OpResult{Halt: types.HaltRevert}, nil
}
