package opcodes

import (
	"math/big"

	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

// OpBLOCKHASH implements the BLOCKHASH opcode (0x40): pushes the hash of the given block number.
// Returns 0 if the block is not among the last 256 or the GetHash function is not set.
func OpBLOCKHASH(e types.Executor) error {
	stack := e.GetStack()
	blockNum, err := stack.Pop()
	if err != nil {
		return err
	}
	var result *big.Int
	getHash := e.GetContext().Block.GetHash
	if getHash != nil {
		hash := getHash(blockNum.Uint64())
		result = new(big.Int).SetBytes(hash[:])
	} else {
		result = new(big.Int)
	}
	if err := stack.Push(result); err != nil {
		return err
	}
	e.SetPC(e.GetPC() + 1)
	return nil
}

// OpCOINBASE implements the COINBASE opcode (0x41): pushes the block beneficiary address.
func OpCOINBASE(e types.Executor) error {
	addr := e.GetContext().Block.Coinbase
	if err := e.GetStack().Push(new(big.Int).SetBytes(addr[:])); err != nil {
		return err
	}
	e.SetPC(e.GetPC() + 1)
	return nil
}

// OpTIMESTAMP implements the TIMESTAMP opcode (0x42): pushes the block timestamp.
func OpTIMESTAMP(e types.Executor) error {
	if err := e.GetStack().Push(new(big.Int).SetUint64(e.GetContext().Block.Timestamp)); err != nil {
		return err
	}
	e.SetPC(e.GetPC() + 1)
	return nil
}

// OpNUMBER implements the NUMBER opcode (0x43): pushes the current block number.
func OpNUMBER(e types.Executor) error {
	if err := e.GetStack().Push(new(big.Int).SetUint64(e.GetContext().Block.Number)); err != nil {
		return err
	}
	e.SetPC(e.GetPC() + 1)
	return nil
}

// OpPREVRANDAO implements the PREVRANDAO opcode (0x44): pushes the previous block's RANDAO mix.
func OpPREVRANDAO(e types.Executor) error {
	r := e.GetContext().Block.PrevRandao
	if err := e.GetStack().Push(new(big.Int).SetBytes(r[:])); err != nil {
		return err
	}
	e.SetPC(e.GetPC() + 1)
	return nil
}

// OpGASLIMIT implements the GASLIMIT opcode (0x45): pushes the block gas limit.
func OpGASLIMIT(e types.Executor) error {
	if err := e.GetStack().Push(new(big.Int).SetUint64(e.GetContext().Block.GasLimit)); err != nil {
		return err
	}
	e.SetPC(e.GetPC() + 1)
	return nil
}

// OpCHAINID implements the CHAINID opcode (0x46): pushes the chain ID.
func OpCHAINID(e types.Executor) error {
	chainID := e.GetContext().Block.ChainID
	if chainID == nil {
		chainID = new(big.Int)
	}
	if err := e.GetStack().Push(new(big.Int).Set(chainID)); err != nil {
		return err
	}
	e.SetPC(e.GetPC() + 1)
	return nil
}

// OpSELFBALANCE implements the SELFBALANCE opcode (0x47): pushes the balance of the executing account.
func OpSELFBALANCE(e types.Executor) error {
	ctx := e.GetContext()
	var balance *big.Int
	if ctx.StateDB != nil {
		balance = ctx.StateDB.GetBalance(ctx.Address)
	} else {
		balance = new(big.Int)
	}
	if err := e.GetStack().Push(balance); err != nil {
		return err
	}
	e.SetPC(e.GetPC() + 1)
	return nil
}

// OpBASEFEE implements the BASEFEE opcode (0x48): pushes the base fee of the current block.
func OpBASEFEE(e types.Executor) error {
	baseFee := e.GetContext().Block.BaseFee
	if baseFee == nil {
		baseFee = new(big.Int)
	}
	if err := e.GetStack().Push(new(big.Int).Set(baseFee)); err != nil {
		return err
	}
	e.SetPC(e.GetPC() + 1)
	return nil
}
