// Copyright 2018 Dmitriy Blokhin. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package consensus

import "time"

// Consensus rule that everything is sorted in lexicographical order on the wire.

// MAXTarget The target is the 32-bytes hash block hashes must be lower than.
var MAXTarget = []byte{0xf, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

const (
	// BlockHashSize size of block hash
	BlockHashSize = 32

	// GrinBase A grin is divisible to 10^9, following the SI prefixes
	GrinBase uint64 = 1E9

	// Milligrin, a thousand of a grin
	MillGrin uint64 = GrinBase / 1000

	// Microgrin, a thousand of a milligrin
	MicroGrin uint64 = MillGrin / 1000

	// Nanogrin, smallest unit, takes a billion to make a grin
	NanoGrin uint64 = 1

	// Reward The block subsidy amount
	Reward uint64 = 60 * GrinBase

	// CoinbaseMaturity Number of blocks before a coinbase matures and can be spent
	CoinbaseMaturity uint64 = 1000

	// MaxBlockCoinbaseOutputs number of max coinbase outputs in a valid block.
	// This is to prevent a miner generating an excessively large "compact block".
	// But we do techincally support blocks with multiple coinbase outputs/kernels.
	MaxBlockCoinbaseOutputs int = 1

	// MaxBlockCoinbaseKernels number of max coinbase kernels in a valid block.
	// This is to prevent a miner generating an excessively large "compact block".
	// But we do techincally support blocks with multiple coinbase outputs/kernels.
	MaxBlockCoinbaseKernels int = 1

	// BlockTimeSec Block interval, in seconds, the network will tune its next_target for. Note
	// that we may reduce this value in the future as we get more data on mining
	// with Cuckoo Cycle, networks improve and block propagation is optimized
	// (adjusting the reward accordingly).
	BlockTimeSec time.Duration = 60

	// ProofSize Cuckoo-cycle proof size (cycle length)
	ProofSize uint32 = 42

	// DefaultSizeshift Default Cuckoo Cycle size shift used for mining and validating.
	DefaultSizeshift uint8 = 30

	// Easiness Default Cuckoo Cycle easiness, high enough to have good likeliness to find
	// a solution.
	Easiness uint32 = 50

	// Default number of blocks in the past when cross-block cut-through will start
	// happening. Needs to be long enough to not overlap with a long reorg.
	// Rational
	// behind the value is the longest bitcoin fork was about 30 blocks, so 5h. We
	// add an order of magnitude to be safe and round to 48h of blocks to make it
	// easier to reason about.
	// TODO: check that on bug with size BlockTimeSec
	CutThroughHorizon uint32 = 48 * 3600 / uint32(BlockTimeSec)

	// Weight of an input when counted against the max block weigth capacity
	BlockInputWeight uint32 = 1

	// Weight of an output when counted against the max block weight capacity
	BlockOutputWeight uint32 = 10

	// Weight of a kernel when counted against the max block weight capacity
	BlockKernelWeight uint32 = 2

	// Total maximum block weight
	MaxBlockWeight uint32 = 80000

	// Fork every 250,000 blocks for first 2 years, simple number and just a
	// little less than 6 months.
	HardForkInterval uint64 = 250000

	// Time window in blocks to calculate block time median
	MedianTimeWindow int = 11

	// Number of blocks used to calculate difficulty adjustments
	DifficultyAdjustWindow int = 23

	// Average time span of the difficulty adjustment window
	BlockTimeWindow time.Duration = time.Duration(DifficultyAdjustWindow) * BlockTimeSec

	// Maximum size time window used for difficulty adjustments
	UpperTimeBound time.Duration = BlockTimeWindow * 4 / 3

	// Minimum size time window used for difficulty adjustments
	LowerTimeBound time.Duration = BlockTimeWindow * 5 / 6
)
