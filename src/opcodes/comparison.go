package opcodes

import (
	"math/big"

	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

// two255 is 2^255, the sign boundary for 256-bit two's complement arithmetic.
var two255 = new(big.Int).Lsh(big.NewInt(1), 255)

// two256 is 2^256, used to convert unsigned values to their signed equivalents.
var two256 = new(big.Int).Lsh(big.NewInt(1), 256)

// toSigned interprets a 256-bit unsigned big.Int as a signed two's complement value.
func toSigned(v *big.Int) *big.Int {
	if v.Cmp(two255) >= 0 {
		return new(big.Int).Sub(v, two256)
	}
	return v
}

func popTwoArgs(stack types.Stack) (a, b *big.Int, err error) {
	a, err = stack.Pop()
	if err != nil {
		return nil, nil, err
	}

	b, err = stack.Pop()
	if err != nil {
		return nil, nil, err
	}

	return a, b, nil
}

// OpEQ implements the EQ opcode, which compares two values for equality.
func OpEQ(e types.Executor) error {
	stack := e.GetStack()

	// EQ takes two arguments from the stack
	a, b, err := popTwoArgs(stack)
	if err != nil {
		return err
	}

	// Compare a and b for equality
	var result uint64
	if a.Cmp(b) == 0 {
		result = 1 // true
	} else {
		result = 0 // false
	}

	// Push the result back onto the stack
	err = stack.Push(new(big.Int).SetUint64(result))
	if err != nil {
		return err
	}

	e.SetPC(e.GetPC() + 1)
	return nil
}

// OpLT implements the LT opcode, which checks if the first value is less than the second.
func OpLT(e types.Executor) error {
	stack := e.GetStack()

	// LT takes two arguments from the stack
	a, b, err := popTwoArgs(stack)
	if err != nil {
		return err
	}

	// Compare a and b for less-than
	var result uint64
	if a.Cmp(b) < 0 {
		result = 1 // true
	} else {
		result = 0 // false
	}

	// Push the result back onto the stack
	err = stack.Push(new(big.Int).SetUint64(result))
	if err != nil {
		return err
	}

	e.SetPC(e.GetPC() + 1)
	return nil
}

// OpGT implements the GT opcode, which checks if the first value is greater than the second.
func OpGT(e types.Executor) error {
	stack := e.GetStack()

	// GT takes two arguments from the stack
	a, b, err := popTwoArgs(stack)
	if err != nil {
		return err
	}

	// Compare a and b for greater-than
	var result uint64
	if a.Cmp(b) > 0 {
		result = 1 // true
	} else {
		result = 0 // false
	}

	// Push the result back onto the stack
	err = stack.Push(new(big.Int).SetUint64(result))
	if err != nil {
		return err
	}

	e.SetPC(e.GetPC() + 1)
	return nil
}

// OpISZERO implements the ISZERO opcode, which checks if a value is zero.
func OpISZERO(e types.Executor) error {
	stack := e.GetStack()

	// ISZERO takes one argument from the stack
	a, err := stack.Pop()
	if err != nil {
		return err
	}

	// Check if a is zero
	var result uint64
	if a.Sign() == 0 {
		result = 1 // true
	} else {
		result = 0 // false
	}

	// Push the result back onto the stack
	err = stack.Push(new(big.Int).SetUint64(result))
	if err != nil {
		return err
	}

	e.SetPC(e.GetPC() + 1)
	return nil
}

// OpSLT implements the SLT opcode, which checks if the first value is less than the second (signed).
func OpSLT(e types.Executor) error {
	stack := e.GetStack()

	// SLT takes two arguments from the stack
	a, b, err := popTwoArgs(stack)
	if err != nil {
		return err
	}

	// Compare a and b for signed less-than (two's complement 256-bit)
	var result uint64
	if toSigned(a).Cmp(toSigned(b)) < 0 {
		result = 1
	}

	// Push the result back onto the stack
	err = stack.Push(new(big.Int).SetUint64(result))
	if err != nil {
		return err
	}

	e.SetPC(e.GetPC() + 1)
	return nil
}

// OpSGT implements the SGT opcode, which checks if the first value is greater than the second (signed).
func OpSGT(e types.Executor) error {
	stack := e.GetStack()

	// SGT takes two arguments from the stack
	a, b, err := popTwoArgs(stack)
	if err != nil {
		return err
	}

	// Compare a and b for signed greater-than (two's complement 256-bit)
	var result uint64
	if toSigned(a).Cmp(toSigned(b)) > 0 {
		result = 1
	}

	// Push the result back onto the stack
	err = stack.Push(new(big.Int).SetUint64(result))
	if err != nil {
		return err
	}

	e.SetPC(e.GetPC() + 1)
	return nil
}
