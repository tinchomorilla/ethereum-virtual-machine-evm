package opcodes

import (
	"math/big"

	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

// mask256 is 2^256 - 1, used for masking results to 256 bits.
var mask256 = new(big.Int).Sub(two256, big.NewInt(1))

// OpAND implements the AND opcode (0x16): bitwise AND of top two stack items.
func OpAND(e types.Executor) (types.OpResult, error) {
	stack := e.GetStack()
	a, b, err := popTwoArgs(stack)
	if err != nil {
		return types.OpResult{}, err
	}
	if err := stack.Push(new(big.Int).And(a, b)); err != nil {
		return types.OpResult{}, err
	}
	e.SetPC(e.GetPC() + 1)
	return types.OpResult{}, nil
}

// OpOR implements the OR opcode (0x17): bitwise OR of top two stack items.
func OpOR(e types.Executor) (types.OpResult, error) {
	stack := e.GetStack()
	a, b, err := popTwoArgs(stack)
	if err != nil {
		return types.OpResult{}, err
	}
	if err := stack.Push(new(big.Int).Or(a, b)); err != nil {
		return types.OpResult{}, err
	}
	e.SetPC(e.GetPC() + 1)
	return types.OpResult{}, nil
}

// OpXOR implements the XOR opcode (0x18): bitwise XOR of top two stack items.
func OpXOR(e types.Executor) (types.OpResult, error) {
	stack := e.GetStack()
	a, b, err := popTwoArgs(stack)
	if err != nil {
		return types.OpResult{}, err
	}
	if err := stack.Push(new(big.Int).Xor(a, b)); err != nil {
		return types.OpResult{}, err
	}
	e.SetPC(e.GetPC() + 1)
	return types.OpResult{}, nil
}

// OpNOT implements the NOT opcode (0x19): bitwise NOT of the top stack item (256-bit).
func OpNOT(e types.Executor) (types.OpResult, error) {
	stack := e.GetStack()
	a, err := stack.Pop()
	if err != nil {
		return types.OpResult{}, err
	}
	// ~a in 256-bit two's complement = (2^256 - 1) XOR a
	if err := stack.Push(new(big.Int).Xor(a, mask256)); err != nil {
		return types.OpResult{}, err
	}
	e.SetPC(e.GetPC() + 1)
	return types.OpResult{}, nil
}

// OpSHL implements the SHL opcode (0x1b): logical shift left.
// Stack input: [shift, value] — pops shift first, then value.
// Result: (value << shift) mod 2^256; 0 if shift >= 256.
func OpSHL(e types.Executor) (types.OpResult, error) {
	stack := e.GetStack()
	shift, value, err := popTwoArgs(stack)
	if err != nil {
		return types.OpResult{}, err
	}
	var result *big.Int
	if shift.Cmp(big.NewInt(256)) >= 0 {
		result = new(big.Int)
	} else {
		result = new(big.Int).Lsh(value, uint(shift.Uint64()))
		result.And(result, mask256)
	}
	if err := stack.Push(result); err != nil {
		return types.OpResult{}, err
	}
	e.SetPC(e.GetPC() + 1)
	return types.OpResult{}, nil
}

// OpSHR implements the SHR opcode (0x1c): logical (unsigned) shift right.
// Stack input: [shift, value] — pops shift first, then value.
// Result: value >> shift (unsigned); 0 if shift >= 256.
func OpSHR(e types.Executor) (types.OpResult, error) {
	stack := e.GetStack()
	shift, value, err := popTwoArgs(stack)
	if err != nil {
		return types.OpResult{}, err
	}
	var result *big.Int
	if shift.Cmp(big.NewInt(256)) >= 0 {
		result = new(big.Int)
	} else {
		result = new(big.Int).Rsh(value, uint(shift.Uint64()))
	}
	if err := stack.Push(result); err != nil {
		return types.OpResult{}, err
	}
	e.SetPC(e.GetPC() + 1)
	return types.OpResult{}, nil
}

// OpSAR implements the SAR opcode (0x1d): arithmetic (signed) shift right.
// Stack input: [shift, value] — pops shift first, then value (256-bit two's complement).
// Result: value >>s shift; fills with sign bit. If shift >= 256: 0 for positive, all-1s for negative.
func OpSAR(e types.Executor) (types.OpResult, error) {
	stack := e.GetStack()
	shift, value, err := popTwoArgs(stack)
	if err != nil {
		return types.OpResult{}, err
	}
	signed := toSigned(value)
	var result *big.Int
	if shift.Cmp(big.NewInt(256)) >= 0 {
		if signed.Sign() < 0 {
			result = new(big.Int).Set(mask256) // -1 in unsigned 256-bit
		} else {
			result = new(big.Int)
		}
	} else {
		result = new(big.Int).Rsh(signed, uint(shift.Uint64()))
		if result.Sign() < 0 {
			result.Add(result, two256)
		}
	}
	if err := stack.Push(result); err != nil {
		return types.OpResult{}, err
	}
	e.SetPC(e.GetPC() + 1)
	return types.OpResult{}, nil
}
