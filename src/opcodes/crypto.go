package opcodes

import (
	"math/big"

	"golang.org/x/crypto/sha3"

	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

// OpKECCAK256 implements the KECCAK256 opcode (0x20).
// Pops offset and size, hashes memory[offset:offset+size] with legacy Keccak-256, and pushes the result.
func OpKECCAK256(e types.Executor) (types.OpResult, error) {
	stack := e.GetStack()
	offset, size, err := popTwoArgs(stack)
	if err != nil {
		return types.OpResult{}, err
	}
	data, err := e.GetMemory().Get(offset.Uint64(), size.Uint64())
	if err != nil {
		return types.OpResult{}, err
	}
	h := sha3.NewLegacyKeccak256()
	h.Write(data)
	digest := h.Sum(nil)
	e.SetPC(e.GetPC() + 1)
	if err := stack.Push(new(big.Int).SetBytes(digest)); err != nil {
		return types.OpResult{}, err
	}
	return types.OpResult{}, nil
}
