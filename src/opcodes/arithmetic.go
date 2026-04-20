package opcodes

import (
	"math/big"

	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

// It is safe to share as long as it is only passed as an argument (read-only)
// to math/big functions and never used as a receiver.
var tt256 = new(big.Int).Lsh(big.NewInt(1), 256)

func OpADD(e types.Executor) error {
	a, err := e.GetStack().Pop()
	if err != nil {
		return err
	}
	b, err := e.GetStack().Pop()
	if err != nil {
		return err
	}

	a.Add(a,b)
	a.Mod(a, tt256)

	e.SetPC(e.GetPC() + 1)
	return e.GetStack().Push(a)
}

func OpSUB(e types.Executor) error {
	a, err := e.GetStack().Pop()
	if err != nil {
		return err
	}
	b, err := e.GetStack().Pop()
	if err != nil {
		return err
	}
	a.Sub(a, b)
	a.Mod(a, tt256)
	e.SetPC(e.GetPC() + 1)
	return e.GetStack().Push(a)
}

func OpMUL(e types.Executor) error {
	a, err := e.GetStack().Pop()
	if err != nil {
		return err
	}
	b, err := e.GetStack().Pop()
	if err != nil {
		return err
	}
	a.Mul(a, b)
	a.Mod(a, tt256)
	e.SetPC(e.GetPC() + 1)
	return e.GetStack().Push(a)
}

func OpDIV(e types.Executor) error {
	a, err := e.GetStack().Pop()
	if err != nil {
		return err
	}
	b, err := e.GetStack().Pop()
	if err != nil {
		return err
	}
	// Division by zero yields 0 per Yellow Paper.
	if b.Sign() == 0 {
		e.SetPC(e.GetPC() + 1)
		return e.GetStack().Push(big.NewInt(0))
	}
	a.Quo(a, b)
	a.Mod(a, tt256)
	e.SetPC(e.GetPC() + 1)
	return e.GetStack().Push(a)
}
