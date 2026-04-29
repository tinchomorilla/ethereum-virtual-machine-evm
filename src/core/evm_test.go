package core_test

import (
	"math/big"
	"testing"

	"golang.org/x/crypto/sha3"

	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/core"
	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/statedb"
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

func TestAND(t *testing.T) {
	// PUSH1 0x0F, PUSH1 0xFF, AND -> 0xFF & 0x0F = 0x0F
	code := []byte{0x60, 0x0f, 0x60, 0xff, 0x16, 0x00}
	evm := core.New(types.ExecutionContext{ByteCode: code})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(big.NewInt(0x0f)) != 0 {
		t.Fatalf("expected 0x0f, got %s", top)
	}
}

func TestOR(t *testing.T) {
	// PUSH1 0x0F, PUSH1 0xF0, OR -> 0xF0 | 0x0F = 0xFF
	code := []byte{0x60, 0x0f, 0x60, 0xf0, 0x17, 0x00}
	evm := core.New(types.ExecutionContext{ByteCode: code})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(big.NewInt(0xff)) != 0 {
		t.Fatalf("expected 0xff, got %s", top)
	}
}

func TestXOR(t *testing.T) {
	// PUSH1 0x0F, PUSH1 0xFF, XOR -> 0xFF ^ 0x0F = 0xF0
	code := []byte{0x60, 0x0f, 0x60, 0xff, 0x18, 0x00}
	evm := core.New(types.ExecutionContext{ByteCode: code})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(big.NewInt(0xf0)) != 0 {
		t.Fatalf("expected 0xf0, got %s", top)
	}
}

func TestNOT(t *testing.T) {
	// PUSH1 0x00, NOT -> all 256 bits set = 2^256 - 1
	code := []byte{0x60, 0x00, 0x19, 0x00}
	evm := core.New(types.ExecutionContext{ByteCode: code})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	// 2^256 - 1
	expected := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	if top.Cmp(expected) != 0 {
		t.Fatalf("expected 2^256-1, got %s", top)
	}
}

func TestSHL(t *testing.T) {
	// PUSH1 0x01 (value), PUSH1 0x01 (shift), SHL -> 1 << 1 = 2
	code := []byte{0x60, 0x01, 0x60, 0x01, 0x1b, 0x00}
	evm := core.New(types.ExecutionContext{ByteCode: code})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(big.NewInt(2)) != 0 {
		t.Fatalf("expected 2, got %s", top)
	}
}

func TestSHR(t *testing.T) {
	// PUSH1 0x04 (value), PUSH1 0x01 (shift), SHR -> 4 >> 1 = 2
	code := []byte{0x60, 0x04, 0x60, 0x01, 0x1c, 0x00}
	evm := core.New(types.ExecutionContext{ByteCode: code})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(big.NewInt(2)) != 0 {
		t.Fatalf("expected 2, got %s", top)
	}
}

func TestSAR(t *testing.T) {
	t.Run("positive: 4 SAR 1 = 2", func(t *testing.T) {
		// PUSH1 0x04 (value), PUSH1 0x01 (shift), SAR -> 4 >>s 1 = 2
		code := []byte{0x60, 0x04, 0x60, 0x01, 0x1d, 0x00}
		evm := core.New(types.ExecutionContext{ByteCode: code})
		if _, err := evm.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		top, _ := evm.State().Stack.Peek(1)
		if top.Cmp(big.NewInt(2)) != 0 {
			t.Fatalf("expected 2, got %s", top)
		}
	})
	t.Run("negative: -4 SAR 1 = -2", func(t *testing.T) {
		// PUSH32 (2^256 - 4 = -4 in two's complement), PUSH1 0x01, SAR -> -2
		neg4 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(4))
		neg4Bytes := make([]byte, 32)
		neg4.FillBytes(neg4Bytes)
		code := append([]byte{0x7f}, neg4Bytes...)
		code = append(code, 0x60, 0x01, 0x1d, 0x00) // PUSH1 1, SAR, STOP
		evm := core.New(types.ExecutionContext{ByteCode: code})
		if _, err := evm.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		top, _ := evm.State().Stack.Peek(1)
		// -2 in unsigned 256-bit = 2^256 - 2
		expected := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(2))
		if top.Cmp(expected) != 0 {
			t.Fatalf("expected 2^256-2 (-2), got %s", top)
		}
	})
}

func TestJUMP(t *testing.T) {
	// Bytecode layout:
	// 0x00: PUSH1 0x04  (destination)
	// 0x02: JUMP
	// 0x03: 0xfe        (invalid — must be skipped)
	// 0x04: JUMPDEST
	// 0x05: PUSH1 0x01  (sentinel: we reached the target)
	// 0x07: STOP
	code := []byte{0x60, 0x04, 0x56, 0xfe, 0x5b, 0x60, 0x01, 0x00}
	evm := core.New(types.ExecutionContext{ByteCode: code})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("expected 1 (reached JUMPDEST), got %s", top)
	}
}

func TestJUMPInvalidDest(t *testing.T) {
	// Attempt to jump into PUSH1 data (position 1 is data, not an opcode).
	// 0x00: PUSH1 0x01  (dest = 1, which is PUSH data, not a valid JUMPDEST)
	// 0x02: JUMP
	code := []byte{0x60, 0x01, 0x56}
	evm := core.New(types.ExecutionContext{ByteCode: code})
	_, err := evm.Run()
	if err == nil {
		t.Fatal("expected error for invalid jump destination, got nil")
	}
}

func TestJUMPI(t *testing.T) {
	t.Run("condition true: jumps to JUMPDEST", func(t *testing.T) {
		// Stack for JUMPI: dest on top (µ's[0]), cond below (µ's[1]).
		// 0x00: PUSH1 0x01   (cond = 1)
		// 0x02: PUSH1 0x08   (dest = 8)
		// 0x04: JUMPI
		// 0x05: PUSH1 0x00   (not reached)
		// 0x07: STOP         (not reached)
		// 0x08: JUMPDEST
		// 0x09: PUSH1 0x01   (sentinel)
		// 0x0b: STOP
		code := []byte{0x60, 0x01, 0x60, 0x08, 0x57, 0x60, 0x00, 0x00, 0x5b, 0x60, 0x01, 0x00}
		evm := core.New(types.ExecutionContext{ByteCode: code})
		if _, err := evm.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		top, _ := evm.State().Stack.Peek(1)
		if top.Cmp(big.NewInt(1)) != 0 {
			t.Fatalf("expected 1 (jumped), got %s", top)
		}
	})
	t.Run("condition false: falls through", func(t *testing.T) {
		// 0x00: PUSH1 0x00   (cond = 0)
		// 0x02: PUSH1 0x08   (dest = 8, unused)
		// 0x04: JUMPI         (no jump)
		// 0x05: PUSH1 0x42   (falls through here)
		// 0x07: STOP
		code := []byte{0x60, 0x00, 0x60, 0x08, 0x57, 0x60, 0x42, 0x00}
		evm := core.New(types.ExecutionContext{ByteCode: code})
		if _, err := evm.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		top, _ := evm.State().Stack.Peek(1)
		if top.Cmp(big.NewInt(0x42)) != 0 {
			t.Fatalf("expected 0x42 (fell through), got %s", top)
		}
	})
}

func TestPC(t *testing.T) {
	// 0x00: PUSH1 0x00  (filler to shift PC opcode to position 2)
	// 0x02: PC           (pushes 2)
	// 0x03: STOP
	code := []byte{0x60, 0x00, 0x58, 0x00}
	evm := core.New(types.ExecutionContext{ByteCode: code})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Stack has two items: 0x00 (from PUSH1) below, and 2 (from PC) on top.
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(big.NewInt(2)) != 0 {
		t.Fatalf("expected PC=2, got %s", top)
	}
}

func TestDUP(t *testing.T) {
	t.Run("DUP1 duplicates top", func(t *testing.T) {
		// PUSH1 0x05, DUP1 -> stack: [5, 5], top = 5, len = 2
		code := []byte{0x60, 0x05, 0x80, 0x00}
		evm := core.New(types.ExecutionContext{ByteCode: code})
		if _, err := evm.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		st := evm.State().Stack
		if st.Len() != 2 {
			t.Fatalf("expected stack len 2, got %d", st.Len())
		}
		top, _ := st.Peek(1)
		if top.Cmp(big.NewInt(5)) != 0 {
			t.Fatalf("expected top 5, got %s", top)
		}
	})
	t.Run("DUP2 duplicates second item", func(t *testing.T) {
		// PUSH1 0x03, PUSH1 0x05, DUP2 -> stack: [3, 5, 3], top = 3
		code := []byte{0x60, 0x03, 0x60, 0x05, 0x81, 0x00}
		evm := core.New(types.ExecutionContext{ByteCode: code})
		if _, err := evm.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		st := evm.State().Stack
		if st.Len() != 3 {
			t.Fatalf("expected stack len 3, got %d", st.Len())
		}
		top, _ := st.Peek(1)
		if top.Cmp(big.NewInt(3)) != 0 {
			t.Fatalf("expected top 3 (copy of second item), got %s", top)
		}
	})
}

func TestSWAP(t *testing.T) {
	t.Run("SWAP1 swaps top two items", func(t *testing.T) {
		// PUSH1 0x03, PUSH1 0x05, SWAP1 -> stack: [5, 3], top = 3
		code := []byte{0x60, 0x03, 0x60, 0x05, 0x90, 0x00}
		evm := core.New(types.ExecutionContext{ByteCode: code})
		if _, err := evm.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		st := evm.State().Stack
		top, _ := st.Peek(1)
		second, _ := st.Peek(2)
		if top.Cmp(big.NewInt(3)) != 0 {
			t.Fatalf("expected top 3, got %s", top)
		}
		if second.Cmp(big.NewInt(5)) != 0 {
			t.Fatalf("expected second 5, got %s", second)
		}
	})
	t.Run("SWAP2 swaps top with third item", func(t *testing.T) {
		// PUSH1 0x01, PUSH1 0x03, PUSH1 0x05, SWAP2 -> top=1, third=5
		code := []byte{0x60, 0x01, 0x60, 0x03, 0x60, 0x05, 0x91, 0x00}
		evm := core.New(types.ExecutionContext{ByteCode: code})
		if _, err := evm.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		st := evm.State().Stack
		top, _ := st.Peek(1)
		third, _ := st.Peek(3)
		if top.Cmp(big.NewInt(1)) != 0 {
			t.Fatalf("expected top 1, got %s", top)
		}
		if third.Cmp(big.NewInt(5)) != 0 {
			t.Fatalf("expected third 5, got %s", third)
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

// ── Arithmetic ────────────────────────────────────────────────────────────────

func TestSDIV(t *testing.T) {
	t.Run("10 / 3 = 3", func(t *testing.T) {
		// PUSH1 3, PUSH1 10, SDIV → 10/3 = 3
		code := []byte{0x60, 0x03, 0x60, 0x0a, 0x05, 0x00}
		evm := core.New(types.ExecutionContext{ByteCode: code})
		if _, err := evm.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		top, _ := evm.State().Stack.Peek(1)
		if top.Cmp(big.NewInt(3)) != 0 {
			t.Fatalf("expected 3, got %s", top)
		}
	})
	t.Run("-10 / 3 = -3", func(t *testing.T) {
		// -10 in 256-bit unsigned = 2^256 - 10
		neg10 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(10))
		neg10b := make([]byte, 32)
		neg10.FillBytes(neg10b)
		code := append([]byte{0x60, 0x03, 0x7f}, neg10b...)
		code = append(code, 0x05, 0x00)
		evm := core.New(types.ExecutionContext{ByteCode: code})
		if _, err := evm.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		top, _ := evm.State().Stack.Peek(1)
		// -3 unsigned = 2^256 - 3
		expected := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(3))
		if top.Cmp(expected) != 0 {
			t.Fatalf("expected 2^256-3 (-3), got %s", top)
		}
	})
	t.Run("divide by zero = 0", func(t *testing.T) {
		code := []byte{0x60, 0x00, 0x60, 0x0a, 0x05, 0x00}
		evm := core.New(types.ExecutionContext{ByteCode: code})
		if _, err := evm.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		top, _ := evm.State().Stack.Peek(1)
		if top.Sign() != 0 {
			t.Fatalf("expected 0, got %s", top)
		}
	})
}

func TestMOD(t *testing.T) {
	// PUSH1 3, PUSH1 10, MOD → 10 % 3 = 1
	code := []byte{0x60, 0x03, 0x60, 0x0a, 0x06, 0x00}
	evm := core.New(types.ExecutionContext{ByteCode: code})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("expected 1, got %s", top)
	}
}

func TestSMOD(t *testing.T) {
	t.Run("10 smod 3 = 1", func(t *testing.T) {
		code := []byte{0x60, 0x03, 0x60, 0x0a, 0x07, 0x00}
		evm := core.New(types.ExecutionContext{ByteCode: code})
		if _, err := evm.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		top, _ := evm.State().Stack.Peek(1)
		if top.Cmp(big.NewInt(1)) != 0 {
			t.Fatalf("expected 1, got %s", top)
		}
	})
	t.Run("-10 smod 3 = -1", func(t *testing.T) {
		neg10 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(10))
		neg10b := make([]byte, 32)
		neg10.FillBytes(neg10b)
		code := append([]byte{0x60, 0x03, 0x7f}, neg10b...)
		code = append(code, 0x07, 0x00)
		evm := core.New(types.ExecutionContext{ByteCode: code})
		if _, err := evm.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		top, _ := evm.State().Stack.Peek(1)
		// -1 unsigned = 2^256 - 1
		expected := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
		if top.Cmp(expected) != 0 {
			t.Fatalf("expected 2^256-1 (-1), got %s", top)
		}
	})
}

func TestADDMOD(t *testing.T) {
	// PUSH1 7, PUSH1 5, PUSH1 10, ADDMOD → (10+5)%7 = 1
	code := []byte{0x60, 0x07, 0x60, 0x05, 0x60, 0x0a, 0x08, 0x00}
	evm := core.New(types.ExecutionContext{ByteCode: code})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("expected 1, got %s", top)
	}
}

func TestMULMOD(t *testing.T) {
	// PUSH1 7, PUSH1 5, PUSH1 4, MULMOD → (4*5)%7 = 6
	code := []byte{0x60, 0x07, 0x60, 0x05, 0x60, 0x04, 0x09, 0x00}
	evm := core.New(types.ExecutionContext{ByteCode: code})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(big.NewInt(6)) != 0 {
		t.Fatalf("expected 6, got %s", top)
	}
}

func TestEXP(t *testing.T) {
	// PUSH1 10, PUSH1 2, EXP → 2^10 = 1024
	code := []byte{0x60, 0x0a, 0x60, 0x02, 0x0a, 0x00}
	evm := core.New(types.ExecutionContext{ByteCode: code})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(big.NewInt(1024)) != 0 {
		t.Fatalf("expected 1024, got %s", top)
	}
}

func TestSIGNEXTEND(t *testing.T) {
	t.Run("sign bit 1: 0xFF with b=0 extends to all 1s", func(t *testing.T) {
		// PUSH1 0xFF, PUSH1 0x00, SIGNEXTEND → 2^256-1
		code := []byte{0x60, 0xff, 0x60, 0x00, 0x0b, 0x00}
		evm := core.New(types.ExecutionContext{ByteCode: code})
		if _, err := evm.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		top, _ := evm.State().Stack.Peek(1)
		expected := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
		if top.Cmp(expected) != 0 {
			t.Fatalf("expected 2^256-1, got %s", top)
		}
	})
	t.Run("sign bit 0: 0x7F with b=0 stays 0x7F", func(t *testing.T) {
		code := []byte{0x60, 0x7f, 0x60, 0x00, 0x0b, 0x00}
		evm := core.New(types.ExecutionContext{ByteCode: code})
		if _, err := evm.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		top, _ := evm.State().Stack.Peek(1)
		if top.Cmp(big.NewInt(0x7f)) != 0 {
			t.Fatalf("expected 0x7f, got %s", top)
		}
	})
}

// ── Crypto ────────────────────────────────────────────────────────────────────

func TestKECCAK256(t *testing.T) {
	// Store 0xFFFF (2 bytes) at memory offset 0, then KECCAK256(offset=0, size=2).
	// Expected: keccak256(0xff, 0xff)
	//
	// PUSH2 0xFFFF, PUSH1 0x00, MSTORE  → stores 0xFFFF right-aligned in 32-byte slot
	// PUSH1 0x02,   PUSH1 0x1e, KECCAK256 → hash bytes at offset 30 (where 0xFFFF sits)
	code := []byte{
		0x61, 0xff, 0xff, // PUSH2 0xFFFF
		0x60, 0x00, // PUSH1 0x00
		0x52,             // MSTORE
		0x60, 0x02,       // PUSH1 2 (size)
		0x60, 0x1e,       // PUSH1 30 (offset: MSTORE places value at bytes 30-31)
		0x20,             // KECCAK256
		0x00,             // STOP
	}
	evm := core.New(types.ExecutionContext{ByteCode: code})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	h := sha3.NewLegacyKeccak256()
	h.Write([]byte{0xff, 0xff})
	expected := new(big.Int).SetBytes(h.Sum(nil))
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(expected) != 0 {
		t.Fatalf("expected %s, got %s", expected, top)
	}
}

// ── Environment (opcodes/env.go) ─────────────────────────────────────────────────────

func TestADDRESS(t *testing.T) {
	var addr types.Address
	addr[19] = 0xab
	evm := core.New(types.ExecutionContext{ByteCode: []byte{0x30, 0x00}, Address: addr})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(new(big.Int).SetBytes(addr[:])) != 0 {
		t.Fatalf("expected %s, got %s", new(big.Int).SetBytes(addr[:]), top)
	}
}

func TestCALLER(t *testing.T) {
	var caller types.Address
	caller[19] = 0xcd
	evm := core.New(types.ExecutionContext{ByteCode: []byte{0x33, 0x00}, Caller: caller})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(new(big.Int).SetBytes(caller[:])) != 0 {
		t.Fatalf("expected %s, got %s", new(big.Int).SetBytes(caller[:]), top)
	}
}

func TestCALLVALUE(t *testing.T) {
	value := big.NewInt(1000)
	evm := core.New(types.ExecutionContext{ByteCode: []byte{0x34, 0x00}, Value: value})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(value) != 0 {
		t.Fatalf("expected %s, got %s", value, top)
	}
}

func TestCALLDATALOAD(t *testing.T) {
	// Input: 32 bytes with 0x42 at byte 0
	input := make([]byte, 32)
	input[0] = 0x42
	// PUSH1 0x00, CALLDATALOAD → loads input[0:32]
	code := []byte{0x60, 0x00, 0x35, 0x00}
	evm := core.New(types.ExecutionContext{ByteCode: code, Input: input})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	expected := new(big.Int).SetBytes(input)
	if top.Cmp(expected) != 0 {
		t.Fatalf("expected %s, got %s", expected, top)
	}
}

func TestCALLDATASIZE(t *testing.T) {
	input := []byte{0x01, 0x02, 0x03}
	code := []byte{0x36, 0x00}
	evm := core.New(types.ExecutionContext{ByteCode: code, Input: input})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(big.NewInt(3)) != 0 {
		t.Fatalf("expected 3, got %s", top)
	}
}

func TestCODESIZE(t *testing.T) {
	// CODESIZE, STOP — code is 2 bytes long
	code := []byte{0x38, 0x00}
	evm := core.New(types.ExecutionContext{ByteCode: code})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(big.NewInt(2)) != 0 {
		t.Fatalf("expected 2, got %s", top)
	}
}

func TestGASPRICE(t *testing.T) {
	price := big.NewInt(20_000_000_000)
	evm := core.New(types.ExecutionContext{ByteCode: []byte{0x3a, 0x00}, GasPrice: price})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(price) != 0 {
		t.Fatalf("expected %s, got %s", price, top)
	}
}

func TestGAS(t *testing.T) {
	// GAS pushes remaining gas after deducting the GAS opcode's own cost (GBase=2).
	code := []byte{0x5a, 0x00}
	evm := core.New(types.ExecutionContext{ByteCode: code})
	before := evm.State().Gas
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	// Gas deducted for the GAS opcode (cost = GBase = 2) before OpGAS runs
	expected := new(big.Int).SetUint64(before - 2)
	if top.Cmp(expected) != 0 {
		t.Fatalf("expected %s, got %s", expected, top)
	}
}

// ── Block info ────────────────────────────────────────────────────────────────

func blockCtx() types.BlockContext {
	return types.BlockContext{
		Coinbase:   types.Address{0xc0},
		Timestamp:  1_700_000_000,
		Number:     19_000_000,
		PrevRandao: types.Hash{0xde, 0xad},
		GasLimit:   30_000_000,
		ChainID:    big.NewInt(1),
		BaseFee:    big.NewInt(10_000_000_000),
		GetHash: func(n uint64) types.Hash {
			var h types.Hash
			h[0] = byte(n)
			return h
		},
	}
}

func TestCOINBASE(t *testing.T) {
	bc := blockCtx()
	evm := core.New(types.ExecutionContext{ByteCode: []byte{0x41, 0x00}, Block: bc})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(new(big.Int).SetBytes(bc.Coinbase[:])) != 0 {
		t.Fatalf("unexpected coinbase: %s", top)
	}
}

func TestTIMESTAMP(t *testing.T) {
	bc := blockCtx()
	evm := core.New(types.ExecutionContext{ByteCode: []byte{0x42, 0x00}, Block: bc})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(new(big.Int).SetUint64(bc.Timestamp)) != 0 {
		t.Fatalf("expected %d, got %s", bc.Timestamp, top)
	}
}

func TestNUMBER(t *testing.T) {
	bc := blockCtx()
	evm := core.New(types.ExecutionContext{ByteCode: []byte{0x43, 0x00}, Block: bc})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(new(big.Int).SetUint64(bc.Number)) != 0 {
		t.Fatalf("expected %d, got %s", bc.Number, top)
	}
}

func TestGASLIMIT(t *testing.T) {
	bc := blockCtx()
	evm := core.New(types.ExecutionContext{ByteCode: []byte{0x45, 0x00}, Block: bc})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(new(big.Int).SetUint64(bc.GasLimit)) != 0 {
		t.Fatalf("expected %d, got %s", bc.GasLimit, top)
	}
}

func TestCHAINID(t *testing.T) {
	bc := blockCtx()
	evm := core.New(types.ExecutionContext{ByteCode: []byte{0x46, 0x00}, Block: bc})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(bc.ChainID) != 0 {
		t.Fatalf("expected %s, got %s", bc.ChainID, top)
	}
}

func TestBASEFEE(t *testing.T) {
	bc := blockCtx()
	evm := core.New(types.ExecutionContext{ByteCode: []byte{0x48, 0x00}, Block: bc})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(bc.BaseFee) != 0 {
		t.Fatalf("expected %s, got %s", bc.BaseFee, top)
	}
}

func TestBLOCKHASH(t *testing.T) {
	bc := blockCtx() // GetHash(n) returns hash with hash[0]=byte(n)
	// PUSH1 0x05, BLOCKHASH → hash of block 5 → hash[0]=5
	code := []byte{0x60, 0x05, 0x40, 0x00}
	evm := core.New(types.ExecutionContext{ByteCode: code, Block: bc})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	var expected types.Hash
	expected[0] = 5
	if top.Cmp(new(big.Int).SetBytes(expected[:])) != 0 {
		t.Fatalf("unexpected blockhash: %s", top)
	}
}

func TestSELFBALANCE(t *testing.T) {
	var addr types.Address
	addr[19] = 0x01
	db := statedb.NewMock()
	db.AddBalance(addr, big.NewInt(5000))
	code := []byte{0x47, 0x00}
	evm := core.New(types.ExecutionContext{ByteCode: code, Address: addr, StateDB: db})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(big.NewInt(5000)) != 0 {
		t.Fatalf("expected 5000, got %s", top)
	}
}


func TestEXTCODESIZE(t *testing.T) {
	var targetAddr types.Address
	targetAddr[19] = 0x99 // Mock address ending in 0x99

	db := statedb.NewMock()
	// Simulating that the target contract has a bytecode of 128 bytes
	db.AddCodeSize(targetAddr, 128) 

	// PUSH1 0x99 (target address), EXTCODESIZE -> pushes 128
	code := []byte{0x60, 0x99, 0x3b, 0x00}
	evm := core.New(types.ExecutionContext{ByteCode: code, StateDB: db})
	
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(big.NewInt(128)) != 0 {
		t.Fatalf("expected 128, got %s", top)
	}
}

func TestSLOAD_ColdSlot(t *testing.T) {
	db := statedb.NewMock()
	var addr types.Address
	addr[19] = 0x01
	code := []byte{
		0x60, 0x01, // PUSH1 0x01 (key)
		0x54,       // SLOAD
		0x00,       // STOP
	}
	evm := core.New(types.ExecutionContext{ByteCode: code, Address: addr, StateDB: db})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Sign() != 0 {
		t.Fatalf("expected zero, got %s", top)
	}
}

func TestSLOAD_PreexistingValue(t *testing.T) {
	db := statedb.NewMock()
	var addr types.Address
	addr[19] = 0x01
	var key types.Hash
	key[31] = 0x01
	var val types.Hash
	val[31] = 0xAB
	db.SetState(addr, key, val)

	code := []byte{
		0x60, 0x01, // PUSH1 0x01 (key)
		0x54,       // SLOAD
		0x00,       // STOP
	}
	evm := core.New(types.ExecutionContext{ByteCode: code, Address: addr, StateDB: db})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	expected := new(big.Int).SetBytes(val[:])
	if top.Cmp(expected) != 0 {
		t.Fatalf("expected %x, got %x", val, top.Bytes())
	}
}

func TestSLOAD_WarmsUpSlot(t *testing.T) {
	db := statedb.NewMock()
	var addr types.Address
	addr[19] = 0x01
	code := []byte{
		0x60, 0x01, // PUSH1 0x01 (key)
		0x54,       // SLOAD
		0x00,       // STOP
	}
	evm := core.New(types.ExecutionContext{ByteCode: code, Address: addr, StateDB: db})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var slot types.Hash
	slot[31] = 0x01
	if !evm.GetAccruedSubstate().IsWarmStorage(addr, slot) {
		t.Fatal("expected slot to be warm after SLOAD")
	}
}

func TestSSTORE_WritesValue(t *testing.T) {
	db := statedb.NewMock()
	var addr types.Address
	addr[19] = 0x01
	code := []byte{
		0x60, 0xff, // PUSH1 0xFF (value)
		0x60, 0x01, // PUSH1 0x01 (key)
		0x55,       // SSTORE
		0x60, 0x01, // PUSH1 0x01 (key)
		0x54,       // SLOAD
		0x00,       // STOP
	}
	evm := core.New(types.ExecutionContext{ByteCode: code, Address: addr, StateDB: db})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(big.NewInt(0xff)) != 0 {
		t.Fatalf("expected 0xff, got %s", top)
	}
}

func TestSSTORE_WarmsUpSlot(t *testing.T) {
	db := statedb.NewMock()
	var addr types.Address
	addr[19] = 0x01
	code := []byte{
		0x60, 0xff, // PUSH1 0xFF (value)
		0x60, 0x01, // PUSH1 0x01 (key)
		0x55,       // SSTORE
		0x00,       // STOP
	}
	evm := core.New(types.ExecutionContext{ByteCode: code, Address: addr, StateDB: db})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var slot types.Hash
	slot[31] = 0x01
	if !evm.GetAccruedSubstate().IsWarmStorage(addr, slot) {
		t.Fatal("expected slot to be warm after SSTORE")
	}
}

func TestSSTORE_OverwriteValue(t *testing.T) {
	db := statedb.NewMock()
	var addr types.Address
	addr[19] = 0x01
	code := []byte{
		0x60, 0xaa, // PUSH1 0xAA (value A)
		0x60, 0x01, // PUSH1 0x01 (key)
		0x55,       // SSTORE
		0x60, 0xbb, // PUSH1 0xBB (value B)
		0x60, 0x01, // PUSH1 0x01 (key)
		0x55,       // SSTORE
		0x60, 0x01, // PUSH1 0x01 (key)
		0x54,       // SLOAD
		0x00,       // STOP
	}
	evm := core.New(types.ExecutionContext{ByteCode: code, Address: addr, StateDB: db})
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(big.NewInt(0xbb)) != 0 {
		t.Fatalf("expected 0xbb, got %s", top)
	}
}

func TestEXTCODEHASH(t *testing.T) {
	var targetAddr types.Address
	targetAddr[19] = 0x88 // Mock address ending in 0x88

	// Create a dummy code hash
	var expectedHash types.Hash
	expectedHash[0] = 0xaa
	expectedHash[31] = 0xbb

	db := statedb.NewMock()
	db.AddCodeHash(targetAddr, expectedHash)

	// PUSH1 0x88 (target address), EXTCODEHASH -> pushes expectedHash
	code := []byte{0x60, 0x88, 0x3f, 0x00}
	evm := core.New(types.ExecutionContext{ByteCode: code, StateDB: db})
	
	if _, err := evm.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	top, _ := evm.State().Stack.Peek(1)
	if top.Cmp(new(big.Int).SetBytes(expectedHash[:])) != 0 {
		t.Fatalf("expected %x, got %x", expectedHash, top.Bytes())
	}
}