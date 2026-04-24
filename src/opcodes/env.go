package opcodes

import (
	"errors"
	"math/big"

	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

var ErrReturnDataOutOfBounds = errors.New("return data out of bounds")

// paddedBytes returns size bytes from src starting at offset, zero-padded if src is shorter.
func paddedBytes(src []byte, offset, size uint64) []byte {
	result := make([]byte, size)
	if offset < uint64(len(src)) {
		copy(result, src[offset:])
	}
	return result
}

// OpADDRESS implements the ADDRESS opcode (0x30): pushes the currently executing account address.
func OpADDRESS(e types.Executor) error {
	addr := e.GetContext().Address
	if err := e.GetStack().Push(new(big.Int).SetBytes(addr[:])); err != nil {
		return err
	}
	e.SetPC(e.GetPC() + 1)
	return nil
}

// OpORIGIN implements the ORIGIN opcode (0x32): pushes the transaction origin address.
func OpORIGIN(e types.Executor) error {
	addr := e.GetContext().Origin
	if err := e.GetStack().Push(new(big.Int).SetBytes(addr[:])); err != nil {
		return err
	}
	e.SetPC(e.GetPC() + 1)
	return nil
}

// OpCALLER implements the CALLER opcode (0x33): pushes the caller address.
func OpCALLER(e types.Executor) error {
	addr := e.GetContext().Caller
	if err := e.GetStack().Push(new(big.Int).SetBytes(addr[:])); err != nil {
		return err
	}
	e.SetPC(e.GetPC() + 1)
	return nil
}

// OpCALLVALUE implements the CALLVALUE opcode (0x34): pushes the value sent with the call (in Wei).
func OpCALLVALUE(e types.Executor) error {
	value := e.GetContext().Value
	if value == nil {
		value = new(big.Int)
	}
	if err := e.GetStack().Push(new(big.Int).Set(value)); err != nil {
		return err
	}
	e.SetPC(e.GetPC() + 1)
	return nil
}

// OpCALLDATALOAD implements the CALLDATALOAD opcode (0x35): pushes 32 bytes of calldata at offset i.
func OpCALLDATALOAD(e types.Executor) error {
	stack := e.GetStack()
	i, err := stack.Pop()
	if err != nil {
		return err
	}
	data := paddedBytes(e.GetContext().Input, i.Uint64(), 32)
	if err := stack.Push(new(big.Int).SetBytes(data)); err != nil {
		return err
	}
	e.SetPC(e.GetPC() + 1)
	return nil
}

// OpCALLDATASIZE implements the CALLDATASIZE opcode (0x36): pushes the byte length of the calldata.
func OpCALLDATASIZE(e types.Executor) error {
	size := uint64(len(e.GetContext().Input))
	if err := e.GetStack().Push(new(big.Int).SetUint64(size)); err != nil {
		return err
	}
	e.SetPC(e.GetPC() + 1)
	return nil
}

// OpCALLDATACOPY implements the CALLDATACOPY opcode (0x37).
// Stack: [destOffset, dataOffset, size] — copies calldata into memory, zero-padding if needed.
func OpCALLDATACOPY(e types.Executor) error {
	stack := e.GetStack()
	destOffset, err := stack.Pop()
	if err != nil {
		return err
	}
	dataOffset, err := stack.Pop()
	if err != nil {
		return err
	}
	size, err := stack.Pop()
	if err != nil {
		return err
	}
	data := paddedBytes(e.GetContext().Input, dataOffset.Uint64(), size.Uint64())
	if err := e.GetMemory().Set(destOffset.Uint64(), size.Uint64(), data); err != nil {
		return err
	}
	e.SetPC(e.GetPC() + 1)
	return nil
}

// OpCODESIZE implements the CODESIZE opcode (0x38): pushes the byte length of the executing code.
func OpCODESIZE(e types.Executor) error {
	size := uint64(len(e.GetCode()))
	if err := e.GetStack().Push(new(big.Int).SetUint64(size)); err != nil {
		return err
	}
	e.SetPC(e.GetPC() + 1)
	return nil
}

// OpCODECOPY implements the CODECOPY opcode (0x39).
// Stack: [destOffset, codeOffset, size] — copies bytecode into memory, zero-padding if needed.
func OpCODECOPY(e types.Executor) error {
	stack := e.GetStack()
	destOffset, err := stack.Pop()
	if err != nil {
		return err
	}
	codeOffset, err := stack.Pop()
	if err != nil {
		return err
	}
	size, err := stack.Pop()
	if err != nil {
		return err
	}
	data := paddedBytes(e.GetCode(), codeOffset.Uint64(), size.Uint64())
	if err := e.GetMemory().Set(destOffset.Uint64(), size.Uint64(), data); err != nil {
		return err
	}
	e.SetPC(e.GetPC() + 1)
	return nil
}

// OpGASPRICE implements the GASPRICE opcode (0x3a): pushes the gas price of the current transaction.
func OpGASPRICE(e types.Executor) error {
	price := e.GetContext().GasPrice
	if price == nil {
		price = new(big.Int)
	}
	if err := e.GetStack().Push(new(big.Int).Set(price)); err != nil {
		return err
	}
	e.SetPC(e.GetPC() + 1)
	return nil
}

// OpRETURNDATASIZE implements the RETURNDATASIZE opcode (0x3d): pushes the size of the last sub-call's return data.
func OpRETURNDATASIZE(e types.Executor) error {
	size := uint64(len(e.GetReturnData()))
	if err := e.GetStack().Push(new(big.Int).SetUint64(size)); err != nil {
		return err
	}
	e.SetPC(e.GetPC() + 1)
	return nil
}

// OpRETURNDATACOPY implements the RETURNDATACOPY opcode (0x3e).
// Stack: [destOffset, dataOffset, size] — copies return data into memory.
// Errors if dataOffset+size exceeds the return data length.
func OpRETURNDATACOPY(e types.Executor) error {
	stack := e.GetStack()
	destOffset, err := stack.Pop()
	if err != nil {
		return err
	}
	dataOffset, err := stack.Pop()
	if err != nil {
		return err
	}
	size, err := stack.Pop()
	if err != nil {
		return err
	}
	returnData := e.GetReturnData()
	end := dataOffset.Uint64() + size.Uint64()
	if end > uint64(len(returnData)) {
		return ErrReturnDataOutOfBounds
	}
	data := returnData[dataOffset.Uint64():end]
	if err := e.GetMemory().Set(destOffset.Uint64(), size.Uint64(), data); err != nil {
		return err
	}
	e.SetPC(e.GetPC() + 1)
	return nil
}

// OpGAS implements the GAS opcode (0x5a): pushes the remaining gas after deducting this instruction's cost.
func OpGAS(e types.Executor) error {
	if err := e.GetStack().Push(new(big.Int).SetUint64(e.GetGas())); err != nil {
		return err
	}
	e.SetPC(e.GetPC() + 1)
	return nil
}
