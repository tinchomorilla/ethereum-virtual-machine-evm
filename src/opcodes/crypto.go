package opcodes

import (
	"math/big"

	"golang.org/x/crypto/sha3"

	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

// OpKECCAK256 implements the KECCAK256 opcode (0x20).
// Pops offset and size, hashes memory[offset:offset+size] with legacy Keccak-256, and pushes the result.
func OpKECCAK256(e types.Executor) error {
	stack := e.GetStack()
	offset, size, err := popTwoArgs(stack)
	if err != nil {
		return err
	}
	data, err := e.GetMemory().Get(offset.Uint64(), size.Uint64())
	if err != nil {
		return err
	}
	h := sha3.NewLegacyKeccak256()
	h.Write(data)
	digest := h.Sum(nil)
	e.SetPC(e.GetPC() + 1)
	return stack.Push(new(big.Int).SetBytes(digest))
}
