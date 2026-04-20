package gas

const (
	GZero          uint64 = 0
	GJumpDest      uint64 = 1
	GBase          uint64 = 2
	GVeryLow       uint64 = 3
	GLow           uint64 = 5
	GMid           uint64 = 8
	GHigh          uint64 = 10
	GExp           uint64 = 10
	GExpByte       uint64 = 10
	GSha3          uint64 = 30
	GSha3Word      uint64 = 6
	GBalance       uint64 = 20
	GSLoad         uint64 = 50
	GSSet          uint64 = 20000
	GSReset        uint64 = 5000
	RSClear        uint64 = 15000
	GSelfDestruct  uint64 = 5000
	GCreate        uint64 = 32000
	GCodeDeposit   uint64 = 200
	GCall          uint64 = 40
	GCallValue     uint64 = 9000
	GCallStipend   uint64 = 2300
	GNewAccount    uint64 = 25000
	GCopy          uint64 = 3
	GMemory        uint64 = 3
	GBlockHash     uint64 = 20
	GExtCode       uint64 = 20
	GTransaction   uint64 = 21000
	GTxCreate      uint64 = 32000
	GTxDataZero    uint64 = 4
	GTxDataNonZero uint64 = 68
	GLog           uint64 = 375
	GLogData       uint64 = 8
	GLogTopic      uint64 = 375
)

// staticCost holds the base gas cost for each opcode (0x00..0xff).
// Opcodes with dynamic behavior store only their static component here.
var staticCost [256]uint64

func init() {
	// 0x00: STOP and arithmetic
	staticCost[0x00] = GZero
	staticCost[0x01] = GVeryLow
	staticCost[0x02] = GLow
	staticCost[0x03] = GVeryLow
	staticCost[0x04] = GLow
	staticCost[0x05] = GLow
	staticCost[0x06] = GLow
	staticCost[0x07] = GLow
	staticCost[0x08] = GMid
	staticCost[0x09] = GMid
	staticCost[0x0a] = GExp // TODO: dynamic cost
	staticCost[0x0b] = GLow

	// 0x10: comparison and bitwise
	for op := byte(0x10); op <= 0x1d; op++ {
		staticCost[op] = GVeryLow
	}

	// 0x20: KECCAK256
	staticCost[0x20] = GSha3 // TODO: dynamic cost

	// 0x30: environmental information
	staticCost[0x30] = GBase
	staticCost[0x31] = GBalance
	staticCost[0x32] = GBase
	staticCost[0x33] = GBase
	staticCost[0x34] = GBase
	staticCost[0x35] = GVeryLow
	staticCost[0x36] = GBase
	staticCost[0x37] = GVeryLow // TODO: dynamic cost
	staticCost[0x38] = GBase
	staticCost[0x39] = GVeryLow // TODO: dynamic cost
	staticCost[0x3a] = GBase
	staticCost[0x3b] = GExtCode
	staticCost[0x3c] = GExtCode // TODO: dynamic cost
	staticCost[0x3d] = GBase
	staticCost[0x3e] = GVeryLow // TODO: dynamic cost
	staticCost[0x3f] = GExtCode

	// 0x40: block information
	staticCost[0x40] = GBlockHash
	staticCost[0x41] = GBase
	staticCost[0x42] = GBase
	staticCost[0x43] = GBase
	staticCost[0x44] = GBase
	staticCost[0x45] = GBase
	staticCost[0x46] = GBase
	staticCost[0x47] = GLow
	staticCost[0x48] = GBase
	staticCost[0x49] = GBase
	staticCost[0x4a] = GBase

	// 0x50: stack, memory, storage and flow
	staticCost[0x50] = GBase
	staticCost[0x51] = GVeryLow // TODO: dynamic cost
	staticCost[0x52] = GVeryLow // TODO: dynamic cost
	staticCost[0x53] = GVeryLow // TODO: dynamic cost
	staticCost[0x54] = GSLoad   // TODO: dynamic cost
	staticCost[0x55] = GSSet    // TODO: dynamic cost
	staticCost[0x56] = GMid
	staticCost[0x57] = GHigh
	staticCost[0x58] = GBase
	staticCost[0x59] = GBase
	staticCost[0x5a] = GBase
	staticCost[0x5b] = GJumpDest

	// 0x60: PUSH1..PUSH32
	for op := byte(0x60); op <= 0x7f; op++ {
		staticCost[op] = GVeryLow
	}

	// 0x80: DUP1..DUP16
	for op := byte(0x80); op <= 0x8f; op++ {
		staticCost[op] = GVeryLow
	}

	// 0x90: SWAP1..SWAP16
	for op := byte(0x90); op <= 0x9f; op++ {
		staticCost[op] = GVeryLow
	}

	// 0xa0: LOG0..LOG4
	for op := byte(0xa0); op <= 0xa4; op++ {
		staticCost[op] = GLog + uint64(op-0xa0)*GLogTopic // TODO: dynamic cost
	}

	// 0xf0: create/call family and halting
	staticCost[0xf0] = GCreate       // TODO: dynamic cost
	staticCost[0xf1] = GCall         // TODO: dynamic cost
	staticCost[0xf2] = GCall         // TODO: dynamic cost
	staticCost[0xf3] = GZero         // TODO: dynamic cost
	staticCost[0xf4] = GCall         // TODO: dynamic cost
	staticCost[0xf5] = GCreate       // TODO: dynamic cost
	staticCost[0xfa] = GCall         // TODO: dynamic cost
	staticCost[0xfd] = GZero         // TODO: dynamic cost
	staticCost[0xff] = GSelfDestruct // TODO: dynamic cost
}
