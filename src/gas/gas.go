package gas

import "github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"

// Cost returns the gas cost for an opcode
// TODO: add dynamic costs (memory expansion, LOG, etc)
func Cost(op types.OpCode, evm any) uint64 {
	_ = evm
	return staticCost[op]
}
