package opcodes

import (
	"math/big"

	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

// Helper to safely pop multiple elements from the stack.
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

// callArgs groups the common arguments for all call operations.
type callArgs struct {
	gasLimit     uint64
	target       types.Address
	value        *big.Int // Will be 0 for calls that do not transfer value
	argsOffset   uint64
	argsSize     uint64
	returnOffset uint64
	returnSize   uint64
}

// execCall is the core engine that executes the sub-EVM and handles memory/snapshots.
// It expects the opcode to have already prepared the arguments and the correct child context.
func execCall(e types.Executor, args callArgs, childCtx types.ExecutionContext, transferValue bool) (types.OpResult, error) {
	stack := e.GetStack()

	// 1. Depth limit check
	if e.GetContext().Depth >= 1024 {
		e.SetPC(e.GetPC() + 1)
		return types.OpResult{}, stack.Push(big.NewInt(0))
	}

	// 2. Balance check (only for standard CALL)
	callerAddr := e.GetContext().Address
	if transferValue && args.value.Sign() > 0 {
		if e.GetContext().ReadOnly {
			return types.OpResult{}, types.ErrWriteProtection
		}
		balance := e.GetContext().StateDB.GetBalance(callerAddr)
		if balance.Cmp(args.value) < 0 {
			e.SetPC(e.GetPC() + 1)
			return types.OpResult{}, stack.Push(big.NewInt(0))
		}
	}

	// 3. Memory expansion (input + return)
	var maxMem uint64
	if args.argsSize > 0 {
		maxMem = args.argsOffset + args.argsSize
	}
	if args.returnSize > 0 {
		if end := args.returnOffset + args.returnSize; end > maxMem {
			maxMem = end
		}
	}
	if maxMem > e.GetMemory().Len() {
		e.GetMemory().Resize(((maxMem + 31) / 32) * 32)
	}

	// 4. Extract input data
	inputData, err := e.GetMemory().Get(args.argsOffset, args.argsSize)
	if err != nil {
		return types.OpResult{}, err
	}
	childCtx.Input = inputData

	// 5. Initial snapshot (MANDATORY for ALL calls)
	snapshotID := e.GetContext().StateDB.Snapshot()

	// 6. ETH transfer (only if applicable)
	if transferValue && args.value.Sign() > 0 {
		e.GetContext().StateDB.SubBalance(callerAddr, args.value)
		e.GetContext().StateDB.AddBalance(childCtx.Address, args.value)
	}

	// 7. Execute sub-context
	childReturnData, haltReason, err := e.RunSubContext(childCtx, args.gasLimit)

	// 8. Failure handling (REVERT or errors)
	if err != nil || haltReason == types.HaltRevert {
		e.GetContext().StateDB.RevertToSnapshot(snapshotID)
		e.SetReturnData(childReturnData)
		e.SetPC(e.GetPC() + 1)
		return types.OpResult{}, stack.Push(big.NewInt(0))
	}

	// 9. Success: store result respecting memory limits
	returnDataToSave := childReturnData
	if uint64(len(returnDataToSave)) > args.returnSize {
		returnDataToSave = returnDataToSave[:args.returnSize]
	}

	e.GetMemory().Set(args.returnOffset, args.returnSize, returnDataToSave)
	e.SetReturnData(childReturnData)
	e.SetPC(e.GetPC() + 1)
	return types.OpResult{}, stack.Push(big.NewInt(1))
}

// OpCALL implements opcode 0xf1. Transfers value and switches context to the target.
func OpCALL(e types.Executor) (types.OpResult, error) {
	args, err := popArgs(7, e.GetStack())
	if err != nil {
		return types.OpResult{}, err
	}

	callParams := callArgs{
		gasLimit:     args[0].Uint64(),
		target:       types.BigIntToAddress(args[1]),
		value:        args[2],
		argsOffset:   args[3].Uint64(),
		argsSize:     args[4].Uint64(),
		returnOffset: args[5].Uint64(),
		returnSize:   args[6].Uint64(),
	}

	parentCtx := e.GetContext()
	childCtx := types.ExecutionContext{
		Address:  callParams.target,
		Origin:   parentCtx.Origin,
		Caller:   parentCtx.Address,
		Value:    callParams.value,
		Depth:    parentCtx.Depth + 1,
		ReadOnly: parentCtx.ReadOnly, // Parent propagates its read-only state
		StateDB:  parentCtx.StateDB,
		ByteCode: parentCtx.StateDB.GetCode(callParams.target),
	}

	return execCall(e, callParams, childCtx, true)
}

// OpSTATICCALL implements opcode 0xfa. Same as CALL, but without value and forces ReadOnly = true.
func OpSTATICCALL(e types.Executor) (types.OpResult, error) {
	args, err := popArgs(6, e.GetStack())
	if err != nil {
		return types.OpResult{}, err
	}

	callParams := callArgs{
		gasLimit:     args[0].Uint64(),
		target:       types.BigIntToAddress(args[1]),
		value:        big.NewInt(0), // STATICCALL has no value slot on the stack
		argsOffset:   args[2].Uint64(),
		argsSize:     args[3].Uint64(),
		returnOffset: args[4].Uint64(),
		returnSize:   args[5].Uint64(),
	}

	parentCtx := e.GetContext()
	childCtx := types.ExecutionContext{
		Address:  callParams.target,
		Origin:   parentCtx.Origin,
		Caller:   parentCtx.Address,
		Value:    big.NewInt(0),
		Depth:    parentCtx.Depth + 1,
		ReadOnly: true, // STATICCALL forces read-only context
		StateDB:  parentCtx.StateDB,
		ByteCode: parentCtx.StateDB.GetCode(callParams.target),
	}

	return execCall(e, callParams, childCtx, false)
}

// OpDELEGATECALL implements opcode 0xf4. Executes external code while preserving
// the parent's Address, Caller and Value.
func OpDELEGATECALL(e types.Executor) (types.OpResult, error) {
	args, err := popArgs(6, e.GetStack())
	if err != nil {
		return types.OpResult{}, err
	}

	callParams := callArgs{
		gasLimit:     args[0].Uint64(),
		target:       types.BigIntToAddress(args[1]),
		value:        big.NewInt(0), // DELEGATECALL does not transfer value
		argsOffset:   args[2].Uint64(),
		argsSize:     args[3].Uint64(),
		returnOffset: args[4].Uint64(),
		returnSize:   args[5].Uint64(),
	}

	parentCtx := e.GetContext()
	childCtx := types.ExecutionContext{
		Address:  parentCtx.Address, // preserve the caller's address
		Caller:   parentCtx.Caller,
		Value:    parentCtx.Value,
		Origin:   parentCtx.Origin,
		Depth:    parentCtx.Depth + 1,
		ReadOnly: parentCtx.ReadOnly,
		StateDB:  parentCtx.StateDB,
		ByteCode: parentCtx.StateDB.GetCode(callParams.target),
	}

	return execCall(e, callParams, childCtx, false)
}

// OpCALLCODE implements opcode 0xf2. Similar to DELEGATECALL but with the target's code and context.
func OpCALLCODE(e types.Executor) (types.OpResult, error) {
	args, err := popArgs(7, e.GetStack())
	if err != nil {
		return types.OpResult{}, err
	}

	callParams := callArgs{
		gasLimit:     args[0].Uint64(),
		target:       types.BigIntToAddress(args[1]), // child address
		value:        args[2],
		argsOffset:   args[3].Uint64(),
		argsSize:     args[4].Uint64(),
		returnOffset: args[5].Uint64(),
		returnSize:   args[6].Uint64(),
	}

	parentCtx := e.GetContext()
	childCtx := types.ExecutionContext{
		Address:  parentCtx.Address, // CALLCODE executes code in the context of the caller (parent)
		Caller:   parentCtx.Address,
		Value:    callParams.value,
		Origin:   parentCtx.Origin,
		Depth:    parentCtx.Depth + 1,
		ReadOnly: parentCtx.ReadOnly,
		StateDB:  parentCtx.StateDB,
		ByteCode: parentCtx.StateDB.GetCode(callParams.target),
	}

	return execCall(e, callParams, childCtx, true)
}
