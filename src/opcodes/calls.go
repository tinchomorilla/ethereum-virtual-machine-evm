package opcodes

import (
	"math/big"

	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

func popArgs(count int, stack types.Stack) ([]*big.Int, error) {
	args := make([]*big.Int, count)
	for i := 0; i < count; i++ {
		var err error
		args[i], err = stack.Pop()
		if err != nil {
			return nil, err
		}
	}
	return args, nil
}

// OpCALL implements the CALL opcode (0xf1).
func OpCALL(e types.Executor) (types.OpResult, error) {
	stack := e.GetStack()
	// Pop the 7 arguments for CALL: gas, to, value, inOffset, inSize, outOffset, outSize.
	args, err := popArgs(7, stack)
	if err != nil {
		return types.OpResult{}, err
	}
	gasLimit, addrInt, value, argsOffset, argsSize, returnOffset, returnSize := args[0], args[1], args[2], args[3], args[4], args[5], args[6]
	addr := types.BigIntToAddress(addrInt)

	// Call depth limit check: EVM does not allow more than 1024 nested calls.
	if e.GetContext().Depth >= 1024 {
		e.SetPC(e.GetPC() + 1)
		return types.OpResult{}, stack.Push(big.NewInt(0)) // 0 means "call failed"
	}

	// Check if the caller has enough balance for the value transfer.
	callerAddr := e.GetContext().Address
	if value.Sign() > 0 {
		balance := e.GetContext().StateDB.GetBalance(callerAddr)
		if balance.Cmp(value) < 0 {
			e.SetPC(e.GetPC() + 1)
			return types.OpResult{}, stack.Push(big.NewInt(0)) // 0 means "call failed"
		}
	}

	var maxMem uint64
	if argsSize.Sign() > 0 {
		maxMem = argsOffset.Uint64() + argsSize.Uint64()
	}
	if returnSize.Sign() > 0 {
		returnEnd := returnOffset.Uint64() + returnSize.Uint64()
		if returnEnd > maxMem {
			maxMem = returnEnd
		}
	}

	// 5. Expand memory if necessary
	if maxMem > e.GetMemory().Len() {
		words := (maxMem + 31) / 32
		e.GetMemory().Resize(words * 32)
	}

	// 6. Extract the input data (calldata) for the child contract
	inputData, err := e.GetMemory().Get(argsOffset.Uint64(), argsSize.Uint64())
	if err != nil {
		return types.OpResult{}, err
	}

	// 1. Create snapshot before any state changes.
	snapshotID := e.GetContext().StateDB.Snapshot()

	// 2. Value transfer.
	if value.Sign() > 0 {
		e.GetContext().StateDB.SubBalance(callerAddr, value)
		e.GetContext().StateDB.AddBalance(addr, value)
	}

	// 3. Build child ExecutionContext.
	childCtx := types.ExecutionContext{
		Address:  addr,
		Origin:   e.GetContext().Origin,
		Caller:   callerAddr,
		Input:    inputData,
		Value:    value,
		Depth:    e.GetContext().Depth + 1,
		StateDB:  e.GetContext().StateDB,
		ByteCode: e.GetContext().StateDB.GetCode(addr),
	}

	// 4. Execute sub-EVM.
	childReturnData, haltReason, err := e.RunSubContext(childCtx, gasLimit.Uint64())

	// 5. Handle result.
	if err != nil || haltReason == types.HaltRevert {
		e.GetContext().StateDB.RevertToSnapshot(snapshotID)
		e.SetReturnData(childReturnData)
		e.SetPC(e.GetPC() + 1)
		return types.OpResult{}, stack.Push(big.NewInt(0))
	}

	// Success: HaltReturn or HaltStop.
	returnDataToSave := childReturnData
	if uint64(len(returnDataToSave)) > returnSize.Uint64() {
		returnDataToSave = returnDataToSave[:returnSize.Uint64()]
	}

	e.GetMemory().Set(returnOffset.Uint64(), returnSize.Uint64(), returnDataToSave)
	e.SetReturnData(childReturnData)
	e.SetPC(e.GetPC() + 1)
	return types.OpResult{}, stack.Push(big.NewInt(1))
}
