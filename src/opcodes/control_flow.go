package opcodes

import (
	"errors"
	"math/big"

	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

var ErrInvalidJumpDest = errors.New("invalid jump destination")

// validJumpDests computes D(c) as defined in Yellow Paper.
// returns the set of valid JUMPDEST positions — positions with opcode 0x5b
// that are actual instructions, not data bytes inside a PUSH.
func ValidJumpDests(code []byte) map[uint64]struct{} {
	dests := make(map[uint64]struct{})
	for i := 0; i < len(code); {
		op := code[i]
		if op == 0x5b {
			dests[uint64(i)] = struct{}{}
		}
		// PUSH1..PUSH32: skip the n immediate data bytes that follow
		if op >= 0x60 && op <= 0x7f {
			i += int(op-0x60) + 2
		} else {
			i++
		}
	}
	return dests
}

// OpJUMP implements the JUMP opcode (0x56): unconditional jump.
// Pops the destination and sets PC to it; errors if dest is not a valid JUMPDEST.
func OpJUMP(e types.Executor) (types.OpResult, error) {
	dest, err := e.GetStack().Pop()
	if err != nil {
		return types.OpResult{}, err
	}
	destVal := dest.Uint64()
	if _, ok := e.GetJumpDests()[destVal]; !ok {
		return types.OpResult{}, ErrInvalidJumpDest
	}
	e.SetPC(destVal)
	return types.OpResult{}, nil
}

// OpJUMPI implements the JUMPI opcode (0x57): conditional jump.
// Stack: µ's[0] = destination (top), µ's[1] = condition.
// Jumps to destination if condition != 0, otherwise advances PC by 1.
func OpJUMPI(e types.Executor) (types.OpResult, error) {
	stack := e.GetStack()
	dest, err := stack.Pop()
	if err != nil {
		return types.OpResult{}, err
	}
	cond, err := stack.Pop()
	if err != nil {
		return types.OpResult{}, err
	}
	if cond.Sign() != 0 {
		destVal := dest.Uint64()
		if _, ok := e.GetJumpDests()[destVal]; !ok {
			return types.OpResult{}, ErrInvalidJumpDest
		}
		e.SetPC(destVal)
	} else {
		e.SetPC(e.GetPC() + 1)
	}
	return types.OpResult{}, nil
}

// OpJUMPDEST implements the JUMPDEST opcode (0x5b): marks a valid jump destination.
// It is a no-op at runtime; its only purpose is to be a legal jump target.
func OpJUMPDEST(e types.Executor) (types.OpResult, error) {
	e.SetPC(e.GetPC() + 1)
	return types.OpResult{}, nil
}

// OpPC implements the PC opcode (0x58): pushes the program counter of this instruction.
func OpPC(e types.Executor) (types.OpResult, error) {
	pc := e.GetPC()
	if err := e.GetStack().Push(new(big.Int).SetUint64(pc)); err != nil {
		return types.OpResult{}, err
	}
	e.SetPC(pc + 1)
	return types.OpResult{}, nil
}
