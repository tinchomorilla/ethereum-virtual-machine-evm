package gas

import "github.com/tinchomorilla/ethereum-virtual-machine-evm/src/types"


// Cost calculates the total gas cost (static + dynamic) for a given opcode.
func Cost(op types.OpCode, e types.Executor) (uint64, error) {
	// Get the base static cost for the opcode
	cost := staticCost[op]

	// If the opcode has dynamic behavior, execute its cost function
	if dynFunc := dynamicCost[op]; dynFunc != nil {
		dynamicPart, err := dynFunc(e)
		if err != nil {
			return 0, err
		}
		cost += dynamicPart
	}

	return cost, nil
}

