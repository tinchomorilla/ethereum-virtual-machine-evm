package stack

import (
	"math/big"
	"testing"
)

// TestStackOverflow verifies that pushing beyond MaxStackDepth returns an error.
func TestStackOverflow(t *testing.T) {
	st := New()
	
	// Fill the stack to its limit
	for i := 0; i < MaxStackDepth; i++ {
		err := st.Push(big.NewInt(int64(i)))
		if err != nil {
			t.Fatalf("unexpected error at push %d: %v", i, err)
		}
	}

	// The 1025th push must fail with overflow
	err := st.Push(big.NewInt(9999))
	if err != ErrStackOverflow {
		t.Errorf("expected ErrStackOverflow, got %v", err)
	}
}

// TestStackUnderflow verifies that popping or peeking an empty stack fails.
func TestStackUnderflow(t *testing.T) {
	st := New()
	
	_, err := st.Pop()
	if err != ErrStackUnderflow {
		t.Errorf("expected ErrStackUnderflow on empty pop, got %v", err)
	}

	_, err = st.Peek(1)
	if err != ErrStackUnderflow {
		t.Errorf("expected ErrStackUnderflow on empty peek, got %v", err)
	}
}

// TestSwap verifies the Swap method exchanges the correct positions.
func TestSwap(t *testing.T) {
	t.Run("Swap(1) exchanges top two items", func(t *testing.T) {
		st := New()
		st.Push(big.NewInt(10))
		st.Push(big.NewInt(20))
		if err := st.Swap(1); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		top, _ := st.Peek(1)
		second, _ := st.Peek(2)
		if top.Cmp(big.NewInt(10)) != 0 {
			t.Errorf("expected top 10, got %s", top)
		}
		if second.Cmp(big.NewInt(20)) != 0 {
			t.Errorf("expected second 20, got %s", second)
		}
	})
	t.Run("Swap(2) exchanges top with third item", func(t *testing.T) {
		st := New()
		st.Push(big.NewInt(1))
		st.Push(big.NewInt(2))
		st.Push(big.NewInt(3))
		if err := st.Swap(2); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		top, _ := st.Peek(1)
		third, _ := st.Peek(3)
		if top.Cmp(big.NewInt(1)) != 0 {
			t.Errorf("expected top 1, got %s", top)
		}
		if third.Cmp(big.NewInt(3)) != 0 {
			t.Errorf("expected third 3, got %s", third)
		}
	})
	t.Run("Swap underflow on insufficient stack", func(t *testing.T) {
		st := New()
		st.Push(big.NewInt(1))
		if err := st.Swap(1); err != ErrStackUnderflow {
			t.Errorf("expected ErrStackUnderflow, got %v", err)
		}
	})
}

// TestStackLogic verifies the correct LIFO behavior and Peek indexing.
func TestStackLogic(t *testing.T) {
	st := New()
	st.Push(big.NewInt(10))
	st.Push(big.NewInt(20))

	if st.Len() != 2 {
		t.Errorf("expected length 2, got %d", st.Len())
	}

	// Peek(1) should return the top element (20)
	val, err := st.Peek(1)
	if err != nil || val.Cmp(big.NewInt(20)) != 0 {
		t.Errorf("expected 20 on Peek(1), got %v", val)
	}

	// Pop should remove and return the top element (20)
	val, _ = st.Pop()
	if val.Cmp(big.NewInt(20)) != 0 {
		t.Errorf("expected pop to return 20, got %v", val)
	}
}