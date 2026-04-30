package gas_test

import (
	"math/big"
	"testing"

	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/gas"
	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/memory"
	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/stack"
	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/substate"
	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

// fakeDB gives tests full, independent control over v0 (committed) and v (current).
type fakeDB struct {
	state     map[types.Hash]types.Hash
	committed map[types.Hash]types.Hash
}

func newFakeDB() *fakeDB {
	return &fakeDB{
		state:     make(map[types.Hash]types.Hash),
		committed: make(map[types.Hash]types.Hash),
	}
}

func (f *fakeDB) GetBalance(types.Address) *big.Int                   { return new(big.Int) }
func (f *fakeDB) AddBalance(types.Address, *big.Int)                  {}
func (f *fakeDB) SubBalance(types.Address, *big.Int)                  {}
func (f *fakeDB) GetCodeSize(types.Address) uint64                    { return 0 }
func (f *fakeDB) GetCodeHash(types.Address) types.Hash                { return types.Hash{} }
func (f *fakeDB) GetState(_ types.Address, key types.Hash) types.Hash { return f.state[key] }
func (f *fakeDB) SetState(_ types.Address, key, val types.Hash)       { f.state[key] = val }
func (f *fakeDB) GetCommittedState(_ types.Address, key types.Hash) types.Hash {
	return f.committed[key]
}

// fakeExec implements types.Executor with real stack, memory, and substate.
type fakeExec struct {
	stk types.Stack
	mem types.Memory
	ctx types.ExecutionContext
	sub *substate.AccruedSubstate
}

func (f *fakeExec) GetStack() types.Stack                     { return f.stk }
func (f *fakeExec) GetMemory() types.Memory                   { return f.mem }
func (f *fakeExec) GetCode() []byte                           { return nil }
func (f *fakeExec) GetPC() uint64                             { return 0 }
func (f *fakeExec) SetPC(uint64)                              {}
func (f *fakeExec) GetJumpDests() map[uint64]struct{}         { return nil }
func (f *fakeExec) GetContext() types.ExecutionContext        { return f.ctx }
func (f *fakeExec) GetGas() uint64                            { return 0 }
func (f *fakeExec) GetReturnData() []byte                     { return nil }
func (f *fakeExec) GetAccruedSubstate() types.AccruedSubstate { return f.sub }

func newExec(db types.StateDB, addr types.Address) *fakeExec {
	return &fakeExec{
		stk: stack.New(),
		mem: memory.New(),
		ctx: types.ExecutionContext{Address: addr, StateDB: db},
		sub: substate.NewAccruedSubstate(),
	}
}

// h returns a 32-byte hash with the last byte set to b.
func h(b byte) types.Hash {
	var hash types.Hash
	hash[31] = b
	return hash
}

// pushKV pushes value then key so that Peek(1)=key and Peek(2)=value.
func pushKV(e *fakeExec, key, value types.Hash) {
	e.stk.Push(new(big.Int).SetBytes(value[:]))
	e.stk.Push(new(big.Int).SetBytes(key[:]))
}

// ── gasSLoad (0x54) ───────────────────────────────────────────────────────────

func TestGasSLoad_ColdSlot(t *testing.T) {
	e := newExec(newFakeDB(), types.Address{})
	slot := h(0x01)
	e.stk.Push(new(big.Int).SetBytes(slot[:]))
	cost, err := gas.Cost(0x54, e)
	if err != nil {
		t.Fatal(err)
	}
	if cost != gas.GColdSload {
		t.Fatalf("expected %d, got %d", gas.GColdSload, cost)
	}
}

func TestGasSLoad_WarmSlot(t *testing.T) {
	var addr types.Address
	slot := h(0x01)
	e := newExec(newFakeDB(), addr)
	e.sub.WarmUpStorage(addr, slot)
	e.stk.Push(new(big.Int).SetBytes(slot[:]))
	cost, err := gas.Cost(0x54, e)
	if err != nil {
		t.Fatal(err)
	}
	if cost != gas.GWarmStorageRead {
		t.Fatalf("expected %d, got %d", gas.GWarmStorageRead, cost)
	}
}

// ── gasSStore (0x55) ──────────────────────────────────────────────────────────

func TestGasSStore_ColdNewSlot(t *testing.T) {
	// cold access, v0=0, v=0, v'≠0 → GColdSload + GSset
	e := newExec(newFakeDB(), types.Address{})
	pushKV(e, h(0x01), h(0xff))
	cost, err := gas.Cost(0x55, e)
	if err != nil {
		t.Fatal(err)
	}
	if cost != gas.GColdSload+gas.GSset {
		t.Fatalf("expected %d, got %d", gas.GColdSload+gas.GSset, cost)
	}
}

func TestGasSStore_WarmNewSlot(t *testing.T) {
	// warm, v0=0, v=0, v'≠0 → GSset
	var addr types.Address
	slot := h(0x01)
	e := newExec(newFakeDB(), addr)
	e.sub.WarmUpStorage(addr, slot)
	pushKV(e, slot, h(0xff))
	cost, err := gas.Cost(0x55, e)
	if err != nil {
		t.Fatal(err)
	}
	if cost != gas.GSset {
		t.Fatalf("expected %d, got %d", gas.GSset, cost)
	}
}

func TestGasSStore_WarmReset(t *testing.T) {
	// warm, v0=A, v=A (first write in tx), v'≠A → GSreset
	var addr types.Address
	slot := h(0x01)
	db := newFakeDB()
	db.state[slot] = h(0x0a)
	db.committed[slot] = h(0x0a)
	e := newExec(db, addr)
	e.sub.WarmUpStorage(addr, slot)
	pushKV(e, slot, h(0xbb))
	cost, err := gas.Cost(0x55, e)
	if err != nil {
		t.Fatal(err)
	}
	if cost != gas.GSreset {
		t.Fatalf("expected %d, got %d", gas.GSreset, cost)
	}
}

func TestGasSStore_WarmNoOp(t *testing.T) {
	// warm, v=v' (no change to current value) → GWarmStorageRead, no refunds
	var addr types.Address
	slot := h(0x01)
	db := newFakeDB()
	db.state[slot] = h(0x42)
	e := newExec(db, addr)
	e.sub.WarmUpStorage(addr, slot)
	pushKV(e, slot, h(0x42))
	cost, err := gas.Cost(0x55, e)
	if err != nil {
		t.Fatal(err)
	}
	if cost != gas.GWarmStorageRead {
		t.Fatalf("cost: expected %d, got %d", gas.GWarmStorageRead, cost)
	}
	if e.sub.Refund != 0 {
		t.Fatalf("refund: expected 0, got %d", e.sub.Refund)
	}
}

func TestGasSStore_SubsequentChange(t *testing.T) {
	// warm, v0=A, v=B (A≠B), v'=C (C≠B) → subsequent write, GWarmStorageRead, no refunds
	var addr types.Address
	slot := h(0x01)
	db := newFakeDB()
	db.committed[slot] = h(0x0a) // v0 = A
	db.state[slot] = h(0x0b)     // v = B ≠ A
	e := newExec(db, addr)
	e.sub.WarmUpStorage(addr, slot)
	pushKV(e, slot, h(0x0c)) // v' = C ≠ B
	cost, err := gas.Cost(0x55, e)
	if err != nil {
		t.Fatal(err)
	}
	if cost != gas.GWarmStorageRead {
		t.Fatalf("cost: expected %d, got %d", gas.GWarmStorageRead, cost)
	}
	if e.sub.Refund != 0 {
		t.Fatalf("refund: expected 0, got %d", e.sub.Refund)
	}
}

func TestGasSStore_ClearRefund(t *testing.T) {
	// warm, v0=A, v=A, v'=0 → cost=GSreset, refund+=RSClear
	var addr types.Address
	slot := h(0x01)
	db := newFakeDB()
	db.state[slot] = h(0x0a)
	db.committed[slot] = h(0x0a)
	e := newExec(db, addr)
	e.sub.WarmUpStorage(addr, slot)
	pushKV(e, slot, types.Hash{}) // v' = 0
	cost, err := gas.Cost(0x55, e)
	if err != nil {
		t.Fatal(err)
	}
	if cost != gas.GSreset {
		t.Fatalf("cost: expected %d, got %d", gas.GSreset, cost)
	}
	if e.sub.Refund != gas.RSClear {
		t.Fatalf("refund: expected %d, got %d", gas.RSClear, e.sub.Refund)
	}
}

func TestGasSStore_RegretRefund(t *testing.T) {
	// warm, v0≠0, v=0, v'≠0 → cost=GWarmStorageRead, refund-=RSClear (undoes earlier clear)
	var addr types.Address
	slot := h(0x01)
	db := newFakeDB()
	db.committed[slot] = h(0x0a)  // v0 = A ≠ 0
	db.state[slot] = types.Hash{} // v = 0 (slot was cleared earlier in this tx)
	e := newExec(db, addr)
	e.sub.WarmUpStorage(addr, slot)
	e.sub.Refund = gas.RSClear // simulate the refund from the earlier clear
	pushKV(e, slot, h(0x05))   // v' ≠ 0
	cost, err := gas.Cost(0x55, e)
	if err != nil {
		t.Fatal(err)
	}
	if cost != gas.GWarmStorageRead {
		t.Fatalf("cost: expected %d, got %d", gas.GWarmStorageRead, cost)
	}
	if e.sub.Refund != 0 {
		t.Fatalf("refund: expected 0 after regret, got %d", e.sub.Refund)
	}
}

func TestGasSStore_RestoreToZeroOriginal(t *testing.T) {
	// warm, v0=0, v≠0, v'=0 → cost=GWarmStorageRead, refund+=RSClear+(GSset-GWarmStorageRead)
	var addr types.Address
	slot := h(0x01)
	db := newFakeDB()
	// committed is zero (default)
	db.state[slot] = h(0x0a) // v = nonzero
	e := newExec(db, addr)
	e.sub.WarmUpStorage(addr, slot)
	pushKV(e, slot, types.Hash{}) // v' = 0 = v0
	cost, err := gas.Cost(0x55, e)
	if err != nil {
		t.Fatal(err)
	}
	if cost != gas.GWarmStorageRead {
		t.Fatalf("cost: expected %d, got %d", gas.GWarmStorageRead, cost)
	}
	wantRefund := gas.RSClear + (gas.GSset - gas.GWarmStorageRead)
	if e.sub.Refund != wantRefund {
		t.Fatalf("refund: expected %d, got %d", wantRefund, e.sub.Refund)
	}
}

func TestGasSStore_RestoreToNonzeroOriginal(t *testing.T) {
	// warm, v0=A, v=B (B≠A), v'=A → cost=GWarmStorageRead, refund+=GSreset-GWarmStorageRead
	var addr types.Address
	slot := h(0x01)
	db := newFakeDB()
	db.committed[slot] = h(0x0a) // v0 = A
	db.state[slot] = h(0x0b)     // v = B ≠ A
	e := newExec(db, addr)
	e.sub.WarmUpStorage(addr, slot)
	pushKV(e, slot, h(0x0a)) // v' = A = v0
	cost, err := gas.Cost(0x55, e)
	if err != nil {
		t.Fatal(err)
	}
	if cost != gas.GWarmStorageRead {
		t.Fatalf("cost: expected %d, got %d", gas.GWarmStorageRead, cost)
	}
	wantRefund := gas.GSreset - gas.GWarmStorageRead
	if e.sub.Refund != wantRefund {
		t.Fatalf("refund: expected %d, got %d", wantRefund, e.sub.Refund)
	}
}

// ── gasMStoreAndMLoad (0x51 / 0x52) ──────────────────────────────────────────

func TestGasMLoad_FreshMemory(t *testing.T) {
	// memory=0, offset=0 → expand to 32 bytes: memoryCost(1)=3, total=GVeryLow+3=6
	e := newExec(nil, types.Address{})
	e.stk.Push(big.NewInt(0))
	cost, err := gas.Cost(0x51, e)
	if err != nil {
		t.Fatal(err)
	}
	if cost != 6 {
		t.Fatalf("expected 6, got %d", cost)
	}
}

func TestGasMLoad_NoExpansion(t *testing.T) {
	// memory already 32 bytes, offset=0 → no expansion, total=GVeryLow=3
	e := newExec(nil, types.Address{})
	e.mem.Resize(32)
	e.stk.Push(big.NewInt(0))
	cost, err := gas.Cost(0x51, e)
	if err != nil {
		t.Fatal(err)
	}
	if cost != gas.GVeryLow {
		t.Fatalf("expected %d, got %d", gas.GVeryLow, cost)
	}
}

func TestGasMLoad_LargeOffset(t *testing.T) {
	// memory=0, offset=64 → expand to 96 bytes (3 words): memoryCost(3)=9, total=GVeryLow+9=12
	e := newExec(nil, types.Address{})
	e.stk.Push(big.NewInt(64))
	cost, err := gas.Cost(0x51, e)
	if err != nil {
		t.Fatal(err)
	}
	if cost != 12 {
		t.Fatalf("expected 12, got %d", cost)
	}
}

// ── gasMStore8 (0x53) ─────────────────────────────────────────────────────────

func TestGasMStore8_FreshMemory(t *testing.T) {
	// memory=0, offset=0 → 1 byte needs 1 word: memoryCost(1)=3, total=GVeryLow+3=6
	e := newExec(nil, types.Address{})
	e.stk.Push(big.NewInt(0))
	cost, err := gas.Cost(0x53, e)
	if err != nil {
		t.Fatal(err)
	}
	if cost != 6 {
		t.Fatalf("expected 6, got %d", cost)
	}
}

func TestGasMStore8_BoundaryExpansion(t *testing.T) {
	// memory=32 bytes (1 word), offset=32 → 33 bytes needs 2 words: expansion=6-3=3, total=6
	e := newExec(nil, types.Address{})
	e.mem.Resize(32)
	e.stk.Push(big.NewInt(32))
	cost, err := gas.Cost(0x53, e)
	if err != nil {
		t.Fatal(err)
	}
	if cost != 6 {
		t.Fatalf("expected 6, got %d", cost)
	}
}

func TestGasMStore8_NoExpansion(t *testing.T) {
	// memory=32 bytes, offset=10 → no expansion, total=GVeryLow=3
	e := newExec(nil, types.Address{})
	e.mem.Resize(32)
	e.stk.Push(big.NewInt(10))
	cost, err := gas.Cost(0x53, e)
	if err != nil {
		t.Fatal(err)
	}
	if cost != gas.GVeryLow {
		t.Fatalf("expected %d, got %d", gas.GVeryLow, cost)
	}
}
