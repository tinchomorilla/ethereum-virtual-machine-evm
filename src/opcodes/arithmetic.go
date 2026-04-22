package opcodes

import (
	"math/big"

	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

func OpADD(e types.Executor) error {
	a, err := e.GetStack().Pop()
	if err != nil {
		return err
	}
	b, err := e.GetStack().Pop()
	if err != nil {
		return err
	}

	a.Add(a, b)
	a.Mod(a, two256)

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
	a.Mod(a, two256)
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
	a.Mod(a, two256)
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
	a.Mod(a, two256)
	e.SetPC(e.GetPC() + 1)
	return e.GetStack().Push(a)
}
