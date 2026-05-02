package opcodes

import (
	"math/big"

	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

// toAddress converts a 256-bit stack value to a 20-byte Ethereum address.
func toAddress(b *big.Int) types.Address {
	var addr types.Address
	bytes := b.Bytes()
	if len(bytes) > 20 {
		copy(addr[:], bytes[len(bytes)-20:])
	} else {
		copy(addr[20-len(bytes):], bytes)
	}
	return addr
}

// OpEXTCODESIZE implements the EXTCODESIZE opcode (0x3b).
// Pops an address from the stack and pushes the byte size of that account's code.
func OpEXTCODESIZE(e types.Executor) (types.OpResult, error) {
	addrInt, err := e.GetStack().Pop()
	if err != nil {
		return types.OpResult{}, err
	}

	addr := toAddress(addrInt)
	var size uint64

	if e.GetContext().StateDB != nil {
		size = e.GetContext().StateDB.GetCodeSize(addr)
	}

	if err := e.GetStack().Push(new(big.Int).SetUint64(size)); err != nil {
		return types.OpResult{}, err
	}
	e.SetPC(e.GetPC() + 1)
	return types.OpResult{}, nil
}

// OpEXTCODEHASH implements the EXTCODEHASH opcode (0x3f).
// Pops an address from the stack and pushes the Keccak-256 hash of that account's code.
func OpEXTCODEHASH(e types.Executor) (types.OpResult, error) {
	addrInt, err := e.GetStack().Pop()
	if err != nil {
		return types.OpResult{}, err
	}

	addr := toAddress(addrInt)
	var hash types.Hash

	if e.GetContext().StateDB != nil {
		hash = e.GetContext().StateDB.GetCodeHash(addr)
	}

	if err := e.GetStack().Push(new(big.Int).SetBytes(hash[:])); err != nil {
		return types.OpResult{}, err
	}
	e.SetPC(e.GetPC() + 1)
	return types.OpResult{}, nil
}
