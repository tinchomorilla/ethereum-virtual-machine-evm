package core

import (
	"errors"

	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/gas"
	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/memory"
	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/opcodes"
	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/stack"
	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

var ErrInvalidOpcode = errors.New("invalid opcode")
var ErrOutOfGas = errors.New("out of gas")

// Executor interface implementation
func (evm *EVM) GetStack() types.Stack   { return evm.state.Stack }
func (evm *EVM) GetMemory() types.Memory { return evm.state.Memory }
func (evm *EVM) GetCode() []byte         { return evm.ctx.ByteCode }
func (evm *EVM) GetPC() uint64           { return evm.state.Pc }
func (evm *EVM) SetPC(pc uint64)         { evm.state.Pc = pc }

// EVM is the Ethereum Virtual Machine.
type EVM struct {
	ctx       types.ExecutionContext
	state     types.MachineState
	jumpTable [256]types.OpFunc
}

// New creates a new EVM instance ready to execute the given context.
func New(ctx types.ExecutionContext) *EVM {
	evm := &EVM{
		ctx: ctx,
		state: types.MachineState{
			Pc:     0,
			Gas:    ^uint64(0), // start with max uint64 gas
			Stack:  stack.New(),
			Memory: memory.New(),
		},
	}
	buildJumpTable(evm)
	return evm
}

func buildJumpTable(evm *EVM) {
	evm.jumpTable[0x00] = opSTOP
	evm.jumpTable[0x01] = opcodes.OpADD
	evm.jumpTable[0x02] = opcodes.OpMUL
	evm.jumpTable[0x03] = opcodes.OpSUB
	evm.jumpTable[0x04] = opcodes.OpDIV
	evm.jumpTable[0x50] = opcodes.OpPOP

	for i := range 32 {
		evm.jumpTable[0x60+i] = opcodes.MakePush(i + 1)
	}
}

// State returns the current machine state
func (evm *EVM) State() *types.MachineState {
	return &evm.state
}

// Run executes the bytecode until STOP, an error occurs or it runs out of gas.
// It returns the output data or an error.
func (evm *EVM) Run() ([]byte, error) {
	code := evm.ctx.ByteCode
	for {
		pc := evm.state.Pc
		if pc >= uint64(len(code)) {
			return nil, ErrInvalidOpcode
		}

		opcode := types.OpCode(code[pc])
		fn := evm.jumpTable[opcode]
		if fn == nil {
			return nil, ErrInvalidOpcode
		}

		cost := gas.Cost(opcode, evm)
		if evm.state.Gas < cost {
			return nil, ErrOutOfGas
		}
		evm.state.Gas -= cost

		if err := fn(evm); err != nil {
			if errors.Is(err, types.ErrStopExecution) {
				return evm.state.ReturnData, nil
			}
			return nil, err
		}
	}
}

func opSTOP(e types.Executor) error {
	return types.ErrStopExecution
}
