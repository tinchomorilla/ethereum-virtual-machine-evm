package core_test

import (
	"math/big"
	"testing"

	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/core"
	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

func TestPUSH1_ADD_STOP(t *testing.T) {
	// PUSH1 0x02, PUSH1 0x03, ADD, STOP
	code := []byte{0x60, 0x02, 0x60, 0x03, 0x01, 0x00}
	evm := core.New(types.ExecutionContext{ByteCode: code})
	_, err := evm.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, err := evm.State().Stack.Peek(1)
	if err != nil {
		t.Fatalf("peek error: %v", err)
	}
	if top.Cmp(big.NewInt(5)) != 0 {
		t.Fatalf("expected 5, got %s", top)
	}
}
