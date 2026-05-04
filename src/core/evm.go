package core

import (
	"errors"

	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/gas"
	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/memory"
	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/opcodes"
	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/stack"
	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/substate"
	"github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"
)

var ErrInvalidOpcode = errors.New("invalid opcode")
var ErrOutOfGas = errors.New("out of gas")

// EVM is the Ethereum Virtual Machine.
type EVM struct {
	ctx             types.ExecutionContext
	state           types.MachineState
	accruedSubstate *substate.AccruedSubstate
	jumpTable       [256]types.OpFunc
	jumpDests       map[uint64]struct{}
}

// New creates a new EVM instance ready to execute the given context.
func New(ctx types.ExecutionContext, initialGas uint64) *EVM {
	evm := &EVM{
		ctx: ctx,
		state: types.MachineState{
			Pc:     0,
			Gas:    initialGas, // start with the provided initial gas
			Stack:  stack.New(),
			Memory: memory.New(),
		},
		accruedSubstate: substate.NewAccruedSubstate(),
		jumpDests:       opcodes.ValidJumpDests(ctx.ByteCode),
	}
	buildJumpTable(evm)
	return evm
}

func buildJumpTable(evm *EVM) {
	evm.jumpTable[0x00] = opcodes.OpSTOP
	evm.jumpTable[0x01] = opcodes.OpADD
	evm.jumpTable[0x02] = opcodes.OpMUL
	evm.jumpTable[0x03] = opcodes.OpSUB
	evm.jumpTable[0x04] = opcodes.OpDIV
	evm.jumpTable[0x05] = opcodes.OpSDIV
	evm.jumpTable[0x06] = opcodes.OpMOD
	evm.jumpTable[0x07] = opcodes.OpSMOD
	evm.jumpTable[0x08] = opcodes.OpADDMOD
	evm.jumpTable[0x09] = opcodes.OpMULMOD
	evm.jumpTable[0x0a] = opcodes.OpEXP
	evm.jumpTable[0x0b] = opcodes.OpSIGNEXTEND
	evm.jumpTable[0x20] = opcodes.OpKECCAK256
	evm.jumpTable[0x30] = opcodes.OpADDRESS
	evm.jumpTable[0x32] = opcodes.OpORIGIN
	evm.jumpTable[0x33] = opcodes.OpCALLER
	evm.jumpTable[0x34] = opcodes.OpCALLVALUE
	evm.jumpTable[0x35] = opcodes.OpCALLDATALOAD
	evm.jumpTable[0x36] = opcodes.OpCALLDATASIZE
	evm.jumpTable[0x37] = opcodes.OpCALLDATACOPY
	evm.jumpTable[0x38] = opcodes.OpCODESIZE
	evm.jumpTable[0x39] = opcodes.OpCODECOPY
	evm.jumpTable[0x3a] = opcodes.OpGASPRICE
	evm.jumpTable[0x3b] = opcodes.OpEXTCODESIZE
	evm.jumpTable[0x3d] = opcodes.OpRETURNDATASIZE
	evm.jumpTable[0x3e] = opcodes.OpRETURNDATACOPY
	evm.jumpTable[0x3f] = opcodes.OpEXTCODEHASH
	evm.jumpTable[0x40] = opcodes.OpBLOCKHASH
	evm.jumpTable[0x41] = opcodes.OpCOINBASE
	evm.jumpTable[0x42] = opcodes.OpTIMESTAMP
	evm.jumpTable[0x43] = opcodes.OpNUMBER
	evm.jumpTable[0x44] = opcodes.OpPREVRANDAO
	evm.jumpTable[0x45] = opcodes.OpGASLIMIT
	evm.jumpTable[0x46] = opcodes.OpCHAINID
	evm.jumpTable[0x47] = opcodes.OpSELFBALANCE
	evm.jumpTable[0x48] = opcodes.OpBASEFEE
	evm.jumpTable[0x5a] = opcodes.OpGAS
	evm.jumpTable[0x10] = opcodes.OpLT
	evm.jumpTable[0x11] = opcodes.OpGT
	evm.jumpTable[0x12] = opcodes.OpSLT
	evm.jumpTable[0x13] = opcodes.OpSGT
	evm.jumpTable[0x14] = opcodes.OpEQ
	evm.jumpTable[0x15] = opcodes.OpISZERO
	evm.jumpTable[0x16] = opcodes.OpAND
	evm.jumpTable[0x17] = opcodes.OpOR
	evm.jumpTable[0x18] = opcodes.OpXOR
	evm.jumpTable[0x19] = opcodes.OpNOT
	evm.jumpTable[0x1b] = opcodes.OpSHL
	evm.jumpTable[0x1c] = opcodes.OpSHR
	evm.jumpTable[0x1d] = opcodes.OpSAR
	evm.jumpTable[0x56] = opcodes.OpJUMP
	evm.jumpTable[0x57] = opcodes.OpJUMPI
	evm.jumpTable[0x58] = opcodes.OpPC
	evm.jumpTable[0x5b] = opcodes.OpJUMPDEST
	evm.jumpTable[0x50] = opcodes.OpPOP
	evm.jumpTable[0x51] = opcodes.OpMLOAD
	evm.jumpTable[0x52] = opcodes.OpMSTORE
	evm.jumpTable[0x53] = opcodes.OpMSTORE8
	evm.jumpTable[0x54] = opcodes.OpSLOAD
	evm.jumpTable[0x55] = opcodes.OpSSTORE
	evm.jumpTable[0x59] = opcodes.OpMSIZE
	evm.jumpTable[0xf1] = opcodes.OpCALL
	evm.jumpTable[0xfd] = opcodes.OpREVERT
	evm.jumpTable[0xf3] = opcodes.OpRETURN

	for i := range 32 {
		evm.jumpTable[0x60+i] = opcodes.MakePush(i + 1)
	}
	for i := range 16 {
		evm.jumpTable[0x80+i] = opcodes.MakeDup(i + 1)
	}
	for i := range 16 {
		evm.jumpTable[0x90+i] = opcodes.MakeSwap(i + 1)
	}
}

// Run executes the bytecode until STOP, an error occurs or it runs out of gas.
// It returns the output data or an error.
func (evm *EVM) Run() ([]byte, types.HaltReason, error) {
	code := evm.ctx.ByteCode
	for {
		pc := evm.state.Pc
		if pc >= uint64(len(code)) {
			evm.state.Gas = 0
			return nil, types.HaltNone, ErrInvalidOpcode
		}

		opcode := types.OpCode(code[pc])
		fn := evm.jumpTable[opcode]
		if fn == nil {
			evm.state.Gas = 0
			return nil, types.HaltNone, ErrInvalidOpcode
		}

		cost, err := gas.Cost(opcode, evm)
		if err != nil {
			evm.state.Gas = 0
			return nil, types.HaltNone, err
		}
		if evm.state.Gas < cost {
			evm.state.Gas = 0
			return nil, types.HaltNone, ErrOutOfGas
		}
		evm.state.Gas -= cost

		res, err := fn(evm)
		if err != nil {
			evm.state.Gas = 0
			return nil, res.Halt, err
		}
		if res.Halt != types.HaltNone {
			return evm.state.ReturnData, res.Halt, nil
		}

	}
}

// RunSubContext instantiates a new child EVM and executes it.
func (evm *EVM) RunSubContext(childCtx types.ExecutionContext, childGas uint64) ([]byte, types.HaltReason, error) {
	childEVM := New(childCtx, childGas)
	return childEVM.Run()
}

// Executor interface implementation
func (evm *EVM) GetStack() types.Stack                     { return evm.state.Stack }
func (evm *EVM) GetMemory() types.Memory                   { return evm.state.Memory }
func (evm *EVM) GetCode() []byte                           { return evm.ctx.ByteCode }
func (evm *EVM) GetPC() uint64                             { return evm.state.Pc }
func (evm *EVM) SetPC(pc uint64)                           { evm.state.Pc = pc }
func (evm *EVM) GetJumpDests() map[uint64]struct{}         { return evm.jumpDests }
func (evm *EVM) GetContext() types.ExecutionContext        { return evm.ctx }
func (evm *EVM) GetGas() uint64                            { return evm.state.Gas }
func (evm *EVM) GetReturnData() []byte                     { return evm.state.ReturnData }
func (evm *EVM) SetReturnData(data []byte)                 { evm.state.ReturnData = data }
func (evm *EVM) GetAccruedSubstate() types.AccruedSubstate { return evm.accruedSubstate }
func (evm *EVM) State() *types.MachineState                { return &evm.state }
