package memory

import (
	"bytes"
	"testing"
)

// TestMemoryResize verifies that the memory expands correctly and pads with zeros.
func TestMemoryResize(t *testing.T) {
	mem := New()
	
	if mem.Len() != 0 {
		t.Errorf("expected initial length 0, got %d", mem.Len())
	}

	// Expand to 32 bytes (1 word)
	mem.Resize(32)
	if mem.Len() != 32 {
		t.Errorf("expected length 32, got %d", mem.Len())
	}

	// Verify it is padded with zeros
	data, _ := mem.Get(0, 32)
	expected := make([]byte, 32)
	if !bytes.Equal(data, expected) {
		t.Errorf("expected zero-padded memory, got %x", data)
	}
}

// TestMemorySetAndGet verifies correct write and read operations.
func TestMemorySetAndGet(t *testing.T) {
	mem := New()
	// Pre-allocate 64 bytes
	mem.Resize(64)

	val := []byte{0x01, 0x02, 0x03}
	err := mem.Set(10, 3, val)
	if err != nil {
		t.Fatalf("unexpected error on Set: %v", err)
	}

	res, err := mem.Get(10, 3)
	if err != nil {
		t.Fatalf("unexpected error on Get: %v", err)
	}

	if !bytes.Equal(val, res) {
		t.Errorf("expected %x, got %x", val, res)
	}
}

// TestMemoryOutOfBounds verifies that accessing unallocated memory fails.
func TestMemoryOutOfBounds(t *testing.T) {
	mem := New()
	mem.Resize(32)

	// Attempt to set beyond allocated size (starts at 30, needs 5 bytes -> ends at 35)
	err := mem.Set(30, 5, []byte{1, 2, 3, 4, 5})
	if err != ErrMemoryOutOfBounds {
		t.Errorf("expected ErrMemoryOutOfBounds on Set, got %v", err)
	}

	// Attempt to get beyond allocated size
	_, err = mem.Get(30, 5)
	if err != ErrMemoryOutOfBounds {
		t.Errorf("expected ErrMemoryOutOfBounds on Get, got %v", err)
	}
}