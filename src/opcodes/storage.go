package opcodes

import (
	"math/big"

	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

func hashToBigInt(h types.Hash) *big.Int {
	return new(big.Int).SetBytes(h[:])
}


// OpSLOAD implements SLOAD (0x54).
func OpSLOAD(e types.Executor) error {
	keyInt, err := e.GetStack().Pop()
	if err != nil {
		return err
	}
	key := types.BigIntToHash(keyInt)
	ctx := e.GetContext()
	addr := ctx.Address

	e.GetAccruedSubstate().WarmUpStorage(addr, key)

	var value types.Hash
	if ctx.StateDB != nil {
		value = ctx.StateDB.GetState(addr, key)
	}

	if err := e.GetStack().Push(hashToBigInt(value)); err != nil {
		return err
	}
	e.SetPC(e.GetPC() + 1)
	return nil
}

// OpSSTORE implements SSTORE (0x55).
func OpSSTORE(e types.Executor) error {
	ctx := e.GetContext()
	if ctx.ReadOnly {
		return types.ErrWriteProtection
	}

	keyInt, err := e.GetStack().Pop()
	if err != nil {
		return err
	}
	valueInt, err := e.GetStack().Pop()
	if err != nil {
		return err
	}

	key := types.BigIntToHash(keyInt)
	value := types.BigIntToHash(valueInt)
	addr := ctx.Address

	e.GetAccruedSubstate().WarmUpStorage(addr, key)

	if ctx.StateDB != nil {
		ctx.StateDB.SetState(addr, key, value)
	}

	e.SetPC(e.GetPC() + 1)
	return nil
}
