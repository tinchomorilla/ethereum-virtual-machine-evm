package opcodes

import (
	"math/big"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
	"golang.org/x/crypto/sha3"
)

// calculateCreateAddress computes the address for a new contract using CREATE.
// Address = Keccak256(RLP([sender_address, sender_nonce]))[12:32]
func calculateCreateAddress(sender types.Address, nonce uint64) (types.Address, error) {
	// RLP encode the sender address and its current nonce
	data := []interface{}{sender[:], nonce}
	rlpBytes, err := rlp.EncodeToBytes(data)
	if err != nil {
		return types.Address{}, err
	}

	// Hash the RLP bytes
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(rlpBytes)
	hash := hasher.Sum(nil)

	// Return the last 20 bytes
	var addr types.Address
	copy(addr[:], hash[12:])
	return addr, nil
}

// calculateCreate2Address computes the address for a new contract using CREATE2.
// Address = Keccak256(0xff || sender_address || salt || Keccak256(init_code))[12:32]
func calculateCreate2Address(sender types.Address, salt *big.Int, initCode []byte) types.Address {
	// Hash the initCode
	codeHasher := sha3.NewLegacyKeccak256()
	codeHasher.Write(initCode)
	codeHash := codeHasher.Sum(nil)

	// Hash 0xff ++ sender ++ salt ++ codeHash
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte{0xff})
	hasher.Write(sender[:])

	saltBytes := make([]byte, 32)
	salt.FillBytes(saltBytes)
	hasher.Write(saltBytes)

	hasher.Write(codeHash)
	hash := hasher.Sum(nil)

	// Return the last 20 bytes
	var addr types.Address
	copy(addr[:], hash[12:])
	return addr
}

// OpCREATE implements opcode 0xf0. It creates a new contract with the given init code.
func OpCREATE(e types.Executor) (types.OpResult, error) {
	stack := e.GetStack()
	memory := e.GetMemory()
	ctx := e.GetContext()

	// 1. Pop arguments: value, offset, size
	args, err := popArgs(3, stack)
	if err != nil {
		return types.OpResult{}, err
	}
	value := args[0]
	initCodeOffset := args[1].Uint64()
	initCodeSize := args[2].Uint64()

	// 2. Write protection check
	if ctx.ReadOnly {
		return types.OpResult{}, types.ErrWriteProtection
	}

	// 3. Depth limit check
	if ctx.Depth >= 1024 {
		e.SetPC(e.GetPC() + 1)
		return types.OpResult{}, stack.Push(big.NewInt(0))
	}

	// 4. Balance check
	callerAddr := ctx.Address
	if ctx.StateDB.GetBalance(callerAddr).Cmp(value) < 0 {
		e.SetPC(e.GetPC() + 1)
		return types.OpResult{}, stack.Push(big.NewInt(0))
	}

	// 5. Memory expansion & Init Code extraction
	var maxMem uint64
	if initCodeSize > 0 {
		maxMem = initCodeOffset + initCodeSize
	}
	if maxMem > memory.Len() {
		memory.Resize(((maxMem + 31) / 32) * 32)
	}

	initCode, err := memory.Get(initCodeOffset, initCodeSize)
	if err != nil {
		return types.OpResult{}, err
	}

	// 6. Calculate new address & Increment caller nonce
	nonce := ctx.StateDB.GetNonce(callerAddr)
	ctx.StateDB.SetNonce(callerAddr, nonce+1)

	newAddr, err := calculateCreateAddress(callerAddr, nonce)
	if err != nil {
		return types.OpResult{}, err
	}

	// 7. Initial snapshot
	snapshotID := ctx.StateDB.Snapshot()

	// 8. ETH transfer and setup of the new account
	if value.Sign() > 0 {
		ctx.StateDB.SubBalance(callerAddr, value)
		ctx.StateDB.AddBalance(newAddr, value)
	}
	ctx.StateDB.SetNonce(newAddr, 1)

	// 9. Setup child context (Execution of Init Code)
	childCtx := types.ExecutionContext{
		Address:  newAddr,
		Origin:   ctx.Origin,
		Caller:   callerAddr,
		Value:    value,
		Input:    nil,      // CREATE receives no calldata
		ByteCode: initCode, // The code to run is the factory code
		Depth:    ctx.Depth + 1,
		ReadOnly: false,
		StateDB:  ctx.StateDB,
	}

	// 10. Execute the sub-context
	childReturnData, haltReason, err := e.RunSubContext(childCtx, e.GetGas())

	// 11. Failure handling
	if err != nil || haltReason == types.HaltRevert {
		ctx.StateDB.RevertToSnapshot(snapshotID)
		e.SetPC(e.GetPC() + 1)
		return types.OpResult{}, stack.Push(big.NewInt(0))
	}

	// 12. Success: store the returned Runtime Code and push new address
	ctx.StateDB.SetCode(newAddr, childReturnData)

	e.SetPC(e.GetPC() + 1)
	return types.OpResult{}, stack.Push(new(big.Int).SetBytes(newAddr[:]))
}

func OpCREATE2(e types.Executor) (types.OpResult, error) {
	stack := e.GetStack()
	memory := e.GetMemory()
	ctx := e.GetContext()

	// 1. Pop arguments: value, offset, size, salt
	args, err := popArgs(4, stack)
	if err != nil {
		return types.OpResult{}, err
	}
	value := args[0]
	initCodeOffset := args[1].Uint64()
	initCodeSize := args[2].Uint64()
	salt := args[3]

	// 2. Write protection check
	if ctx.ReadOnly {
		return types.OpResult{}, types.ErrWriteProtection
	}

	// 3. Depth limit check
	if ctx.Depth >= 1024 {
		e.SetPC(e.GetPC() + 1)
		return types.OpResult{}, stack.Push(big.NewInt(0))
	}

	// 4. Balance check
	callerAddr := ctx.Address
	if ctx.StateDB.GetBalance(callerAddr).Cmp(value) < 0 {
		e.SetPC(e.GetPC() + 1)
		return types.OpResult{}, stack.Push(big.NewInt(0))
	}

	// 5. Memory expansion & Init Code extraction
	var maxMem uint64
	if initCodeSize > 0 {
		maxMem = initCodeOffset + initCodeSize
	}
	if maxMem > memory.Len() {
		memory.Resize(((maxMem + 31) / 32) * 32)
	}

	initCode, err := memory.Get(initCodeOffset, initCodeSize)
	if err != nil {
		return types.OpResult{}, err
	}

	// 6. Calculate new address (CREATE2 is deterministic and doesn't use nonce)
	newAddr := calculateCreate2Address(callerAddr, salt, initCode)

	// 7. Initial snapshot
	snapshotID := ctx.StateDB.Snapshot()

	// 8. ETH transfer and setup of the new account
	if value.Sign() > 0 {
		ctx.StateDB.SubBalance(callerAddr, value)
		ctx.StateDB.AddBalance(newAddr, value)
	}
	ctx.StateDB.SetNonce(newAddr, 1)

	// 9. Setup child context (Execution of Init Code)
	childCtx := types.ExecutionContext{
		Address:  newAddr,
		Origin:   ctx.Origin,
		Caller:   callerAddr,
		Value:    value,
		Input:    nil,      // CREATE2 receives no calldata
		ByteCode: initCode, // The code to run is the factory code
		Depth:    ctx.Depth + 1,
		ReadOnly: false,
		StateDB:  ctx.StateDB,
	}

	// 10. Execute the sub-context
	childReturnData, haltReason, err := e.RunSubContext(childCtx, e.GetGas())

	// 11. Failure handling
	if err != nil || haltReason == types.HaltRevert {
		ctx.StateDB.RevertToSnapshot(snapshotID)
		e.SetPC(e.GetPC() + 1)
		return types.OpResult{}, stack.Push(big.NewInt(0))
	}

	// 12. Success: store the returned Runtime Code and push new address
	ctx.StateDB.SetCode(newAddr, childReturnData)

	e.SetPC(e.GetPC() + 1)
	return types.OpResult{}, stack.Push(new(big.Int).SetBytes(newAddr[:]))
}
