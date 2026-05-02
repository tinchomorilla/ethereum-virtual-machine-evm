package opcodes

import (
	"math/big"

	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

func OpADD(e types.Executor) (types.OpResult, error) {
	a, err := e.GetStack().Pop()
	if err != nil {
		return types.OpResult{}, err
	}
	b, err := e.GetStack().Pop()
	if err != nil {
		return types.OpResult{}, err
	}

	a.Add(a, b)
	a.Mod(a, two256)

	e.SetPC(e.GetPC() + 1)
	if err := e.GetStack().Push(a); err != nil {
		return types.OpResult{}, err
	}
	return types.OpResult{}, nil
}

func OpSUB(e types.Executor) (types.OpResult, error) {
	a, err := e.GetStack().Pop()
	if err != nil {
		return types.OpResult{}, err
	}
	b, err := e.GetStack().Pop()
	if err != nil {
		return types.OpResult{}, err
	}
	a.Sub(a, b)
	a.Mod(a, two256)
	e.SetPC(e.GetPC() + 1)
	if err := e.GetStack().Push(a); err != nil {
		return types.OpResult{}, err
	}
	return types.OpResult{}, nil
}

func OpMUL(e types.Executor) (types.OpResult, error) {
	a, err := e.GetStack().Pop()
	if err != nil {
		return types.OpResult{}, err
	}
	b, err := e.GetStack().Pop()
	if err != nil {
		return types.OpResult{}, err
	}
	a.Mul(a, b)
	a.Mod(a, two256)
	e.SetPC(e.GetPC() + 1)
	if err := e.GetStack().Push(a); err != nil {
		return types.OpResult{}, err
	}
	return types.OpResult{}, nil
}

func OpDIV(e types.Executor) (types.OpResult, error) {
	a, err := e.GetStack().Pop()
	if err != nil {
		return types.OpResult{}, err
	}
	b, err := e.GetStack().Pop()
	if err != nil {
		return types.OpResult{}, err
	}
	// Division by zero yields 0 per Yellow Paper.
	if b.Sign() == 0 {
		e.SetPC(e.GetPC() + 1)
		if err := e.GetStack().Push(big.NewInt(0)); err != nil {
			return types.OpResult{}, err
		}
		return types.OpResult{}, nil
	}
	a.Quo(a, b)
	a.Mod(a, two256)
	e.SetPC(e.GetPC() + 1)
	if err := e.GetStack().Push(a); err != nil {
		return types.OpResult{}, err
	}
	return types.OpResult{}, nil
}

// OpSDIV implements the SDIV opcode (0x05): signed integer division, truncated toward zero.
func OpSDIV(e types.Executor) (types.OpResult, error) {
	a, b, err := popTwoArgs(e.GetStack())
	if err != nil {
		return types.OpResult{}, err
	}
	var result *big.Int
	if b.Sign() == 0 {
		result = new(big.Int)
	} else {
		sa, sb := toSigned(a), toSigned(b)
		result = new(big.Int).Quo(sa, sb)
		if result.Sign() < 0 {
			result.Add(result, two256)
		}
	}
	e.SetPC(e.GetPC() + 1)
	if err := e.GetStack().Push(result); err != nil {
		return types.OpResult{}, err
	}
	return types.OpResult{}, nil
}

// OpMOD implements the MOD opcode (0x06): unsigned modulo.
func OpMOD(e types.Executor) (types.OpResult, error) {
	a, b, err := popTwoArgs(e.GetStack())
	if err != nil {
		return types.OpResult{}, err
	}
	var result *big.Int
	if b.Sign() == 0 {
		result = new(big.Int)
	} else {
		result = new(big.Int).Mod(a, b)
	}
	e.SetPC(e.GetPC() + 1)
	if err := e.GetStack().Push(result); err != nil {
		return types.OpResult{}, err
	}
	return types.OpResult{}, nil
}

// OpSMOD implements the SMOD opcode (0x07): signed modulo (result carries the sign of the dividend).
func OpSMOD(e types.Executor) (types.OpResult, error) {
	a, b, err := popTwoArgs(e.GetStack())
	if err != nil {
		return types.OpResult{}, err
	}
	var result *big.Int
	if b.Sign() == 0 {
		result = new(big.Int)
	} else {
		sa, sb := toSigned(a), toSigned(b)
		result = new(big.Int).Rem(sa, sb)
		if result.Sign() < 0 {
			result.Add(result, two256)
		}
	}
	e.SetPC(e.GetPC() + 1)
	if err := e.GetStack().Push(result); err != nil {
		return types.OpResult{}, err
	}
	return types.OpResult{}, nil
}

// OpADDMOD implements the ADDMOD opcode (0x08): (a + b) % N, with addition done without 256-bit overflow.
func OpADDMOD(e types.Executor) (types.OpResult, error) {
	stack := e.GetStack()
	a, err := stack.Pop()
	if err != nil {
		return types.OpResult{}, err
	}
	b, err := stack.Pop()
	if err != nil {
		return types.OpResult{}, err
	}
	n, err := stack.Pop()
	if err != nil {
		return types.OpResult{}, err
	}
	var result *big.Int
	if n.Sign() == 0 {
		result = new(big.Int)
	} else {
		result = new(big.Int).Add(a, b)
		result.Mod(result, n)
	}
	e.SetPC(e.GetPC() + 1)
	if err := stack.Push(result); err != nil {
		return types.OpResult{}, err
	}
	return types.OpResult{}, nil
}

// OpMULMOD implements the MULMOD opcode (0x09): (a * b) % N, with multiplication done without 256-bit overflow.
func OpMULMOD(e types.Executor) (types.OpResult, error) {
	stack := e.GetStack()
	a, err := stack.Pop()
	if err != nil {
		return types.OpResult{}, err
	}
	b, err := stack.Pop()
	if err != nil {
		return types.OpResult{}, err
	}
	n, err := stack.Pop()
	if err != nil {
		return types.OpResult{}, err
	}
	var result *big.Int
	if n.Sign() == 0 {
		result = new(big.Int)
	} else {
		result = new(big.Int).Mul(a, b)
		result.Mod(result, n)
	}
	e.SetPC(e.GetPC() + 1)
	if err := stack.Push(result); err != nil {
		return types.OpResult{}, err
	}
	return types.OpResult{}, nil
}

// OpEXP implements the EXP opcode (0x0a): a^b mod 2^256.
func OpEXP(e types.Executor) (types.OpResult, error) {
	a, b, err := popTwoArgs(e.GetStack())
	if err != nil {
		return types.OpResult{}, err
	}
	result := new(big.Int).Exp(a, b, two256)
	e.SetPC(e.GetPC() + 1)
	if err := e.GetStack().Push(result); err != nil {
		return types.OpResult{}, err
	}
	return types.OpResult{}, nil
}

// OpSIGNEXTEND implements the SIGNEXTEND opcode (0x0b).
// b is the byte index (0 = LSB). Sign-extends x from bit (b*8+7) through the full 256-bit word.
func OpSIGNEXTEND(e types.Executor) (types.OpResult, error) {
	b, x, err := popTwoArgs(e.GetStack())
	if err != nil {
		return types.OpResult{}, err
	}
	var result *big.Int
	if b.Cmp(big.NewInt(31)) >= 0 {
		result = new(big.Int).Set(x)
	} else {
		signBitPos := uint(b.Uint64())*8 + 7
		// mask covers bits 0..signBitPos inclusive
		mask := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), signBitPos+1), big.NewInt(1))
		if x.Bit(int(signBitPos)) == 1 {
			// Sign bit is 1: fill all bits above signBitPos with 1s
			highBits := new(big.Int).Xor(mask256, mask)
			result = new(big.Int).Or(new(big.Int).And(x, mask), highBits)
		} else {
			// Sign bit is 0: clear all bits above signBitPos
			result = new(big.Int).And(x, mask)
		}
	}
	e.SetPC(e.GetPC() + 1)
	if err := e.GetStack().Push(result); err != nil {
		return types.OpResult{}, err
	}
	return types.OpResult{}, nil
}
