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

func TestMemoryStoreAndLoad(t *testing.T) {
	// PUSH1 0xFF (Value to store: 255)
	// PUSH1 0x00 (Offset 0)
	// MSTORE     (Store in memory)
	// PUSH1 0x00 (Offset 0)
	// MLOAD      (Load from memory to stack)
	// STOP
	code := []byte{
		0x60, 0xff,
		0x60, 0x00,
		0x52,
		0x60, 0x00,
		0x51,
		0x00,
	}

	evm := core.New(types.ExecutionContext{ByteCode: code})
	_, err := evm.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Stack should have 0xFF on top after MLOAD
	top, err := evm.State().Stack.Peek(1)
	if err != nil {
		t.Fatalf("peek error: %v", err)
	}
	if top.Cmp(big.NewInt(0xff)) != 0 {
		t.Fatalf("expected 255, got %s", top)
	}
}

func TestComparisonOpcodes(t *testing.T) {
	// PUSH1 0x0A (10)
	// PUSH1 0x0A (10)
	// EQ         (10 == 10 ? -> 1)
	// STOP
	code := []byte{
		0x60, 0x0a,
		0x60, 0x0a,
		0x14,
		0x00,
	}

	evm := core.New(types.ExecutionContext{ByteCode: code})
	_, err := evm.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Stack should have 1 (true) on top after EQ
	top, err := evm.State().Stack.Peek(1)
	if err != nil {
		t.Fatalf("peek error: %v", err)
	}
	if top.Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("expected 1 (true), got %s", top)
	}
}

func TestLT(t *testing.T) {
	// PUSH1 5, PUSH1 3, LT -> a=3, b=5 -> 3 < 5 -> 1
	code := []byte{0x60, 0x05, 0x60, 0x03, 0x10, 0x00}
	evm := core.New(types.ExecutionContext{ByteCode: code})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("expected 1, got %s", top)
	}
}

func TestGT(t *testing.T) {
	// PUSH1 3, PUSH1 5, GT -> a=5, b=3 -> 5 > 3 -> 1
	code := []byte{0x60, 0x03, 0x60, 0x05, 0x11, 0x00}
	evm := core.New(types.ExecutionContext{ByteCode: code})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("expected 1, got %s", top)
	}
}

func TestISZERO(t *testing.T) {
	// PUSH1 0, ISZERO -> 1; then PUSH1 7, ISZERO -> 0
	t.Run("zero", func(t *testing.T) {
		code := []byte{0x60, 0x00, 0x15, 0x00}
		evm := core.New(types.ExecutionContext{ByteCode: code})
		if _, err := evm.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		top, _ := evm.State().Stack.Peek(1)
		if top.Cmp(big.NewInt(1)) != 0 {
			t.Fatalf("expected 1, got %s", top)
		}
	})
	t.Run("nonzero", func(t *testing.T) {
		code := []byte{0x60, 0x07, 0x15, 0x00}
		evm := core.New(types.ExecutionContext{ByteCode: code})
		if _, err := evm.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		top, _ := evm.State().Stack.Peek(1)
		if top.Cmp(big.NewInt(0)) != 0 {
			t.Fatalf("expected 0, got %s", top)
		}
	})
}

func TestMSTORE8(t *testing.T) {
	// PUSH1 0xAB, PUSH1 0x00, MSTORE8 -> memory[0] = 0xAB
	// PUSH1 0x00, MLOAD -> loads 32 bytes from offset 0: 0xAB at MSB
	code := []byte{0x60, 0xab, 0x60, 0x00, 0x53, 0x60, 0x00, 0x51, 0x00}
	evm := core.New(types.ExecutionContext{ByteCode: code})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	expected := new(big.Int).Lsh(big.NewInt(0xab), 248) // 0xAB at byte 0 of a 32-byte word
	if top.Cmp(expected) != 0 {
		t.Fatalf("expected %s, got %s", expected, top)
	}
}

func TestSLT(t *testing.T) {
	t.Run("positive: 3 SLT 5 = 1", func(t *testing.T) {
		// PUSH1 5, PUSH1 3, SLT -> a=3, b=5 -> 3 <s 5 -> 1
		code := []byte{0x60, 0x05, 0x60, 0x03, 0x12, 0x00}
		evm := core.New(types.ExecutionContext{ByteCode: code})
		if _, err := evm.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		top, _ := evm.State().Stack.Peek(1)
		if top.Cmp(big.NewInt(1)) != 0 {
			t.Fatalf("expected 1, got %s", top)
		}
	})
	t.Run("signed: -1 SLT 0 = 1", func(t *testing.T) {
		// PUSH1 0, PUSH32 (2^256-1 = -1), SLT -> a=-1, b=0 -> -1 <s 0 -> 1
		neg1 := make([]byte, 32)
		for i := range neg1 {
			neg1[i] = 0xff
		}
		code := append([]byte{0x60, 0x00, 0x7f}, neg1...)
		code = append(code, 0x12, 0x00) // SLT, STOP
		evm := core.New(types.ExecutionContext{ByteCode: code})
		if _, err := evm.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		top, _ := evm.State().Stack.Peek(1)
		if top.Cmp(big.NewInt(1)) != 0 {
			t.Fatalf("expected 1, got %s", top)
		}
	})
}

func TestSGT(t *testing.T) {
	t.Run("positive: 5 SGT 3 = 1", func(t *testing.T) {
		// PUSH1 3, PUSH1 5, SGT -> a=5, b=3 -> 5 >s 3 -> 1
		code := []byte{0x60, 0x03, 0x60, 0x05, 0x13, 0x00}
		evm := core.New(types.ExecutionContext{ByteCode: code})
		if _, err := evm.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		top, _ := evm.State().Stack.Peek(1)
		if top.Cmp(big.NewInt(1)) != 0 {
			t.Fatalf("expected 1, got %s", top)
		}
	})
	t.Run("signed: 0 SGT -1 = 1", func(t *testing.T) {
		// PUSH32 (2^256-1 = -1), PUSH1 0, SGT -> a=0, b=-1 -> 0 >s -1 -> 1
		neg1 := make([]byte, 32)
		for i := range neg1 {
			neg1[i] = 0xff
		}
		code := append([]byte{0x7f}, neg1...)
		code = append(code, 0x60, 0x00, 0x13, 0x00) // PUSH1 0, SGT, STOP
		evm := core.New(types.ExecutionContext{ByteCode: code})
		if _, err := evm.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		top, _ := evm.State().Stack.Peek(1)
		if top.Cmp(big.NewInt(1)) != 0 {
			t.Fatalf("expected 1, got %s", top)
		}
	})
}

func TestMSIZE(t *testing.T) {
	// PUSH1 0x01, PUSH1 0x00, MSTORE (expands memory to 32 bytes), MSIZE -> 32
	code := []byte{0x60, 0x01, 0x60, 0x00, 0x52, 0x59, 0x00}
	evm := core.New(types.ExecutionContext{ByteCode: code})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(big.NewInt(32)) != 0 {
		t.Fatalf("expected 32, got %s", top)
	}
}
