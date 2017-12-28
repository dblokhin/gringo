package consensus

// The target is the 32-bytes hash block hashes must be lower than.
var MAX_TARGET = [8]uint8{0xf, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

const (
	// A grin is divisible to 10^9, following the SI prefixes
	GRIN_BASE uint64 = 1E9

	// Milligrin, a thousand of a grin
	MILLI_GRIN uint64 = GRIN_BASE / 1000

	// Microgrin, a thousand of a milligrin
	MICRO_GRIN uint64 = MILLI_GRIN / 1000

	// Nanogrin, smallest unit, takes a billion to make a grin
	NANO_GRIN uint64 = 1

	// The block subsidy amount
	REWARD uint64 = 50 * GRIN_BASE

	// Number of blocks before a coinbase matures and can be spent
	COINBASE_MATURITY uint64 = 1000

	// Block interval, in seconds, the network will tune its next_target for. Note
	// that we may reduce this value in the future as we get more data on mining
	// with Cuckoo Cycle, networks improve and block propagation is optimized
	// (adjusting the reward accordingly).
	BLOCK_TIME_SEC uint64 = 60

	// Cuckoo-cycle proof size (cycle length)
	PROOFSIZE uint32 = 42

	// Default Cuckoo Cycle size shift used for mining and validating.
	DEFAULT_SIZESHIFT uint8 = 30

	// Default Cuckoo Cycle easiness, high enough to have good likeliness to find
	// a solution.
	EASINESS uint32 = 50

	// Default number of blocks in the past when cross-block cut-through will start
	// happening. Needs to be long enough to not overlap with a long reorg.
	// Rational
	// behind the value is the longest bitcoin fork was about 30 blocks, so 5h. We
	// add an order of magnitude to be safe and round to 48h of blocks to make it
	// easier to reason about.
	CUT_THROUGH_HORIZON uint32 = 48 * 3600 / uint32(BLOCK_TIME_SEC)

	// The maximum size we're willing to accept for any message. Enforced by the
	// peer-to-peer networking layer only for DoS protection.
	MAX_MSG_LEN uint64 = 20000000

	// Weight of an input when counted against the max block weigth capacity
	BLOCK_INPUT_WEIGHT uint32 = 1

	// Weight of an output when counted against the max block weight capacity
	BLOCK_OUTPUT_WEIGHT uint32 = 10

	// Weight of a kernel when counted against the max block weight capacity
	BLOCK_KERNEL_WEIGHT uint32 = 2

	// Total maximum block weight
	MAX_BLOCK_WEIGHT uint32 = 80000

	// Fork every 250,000 blocks for first 2 years, simple number and just a
	// little less than 6 months.
	HARD_FORK_INTERVAL uint64 = 250000

	// The minimum mining difficulty we'll allow
	MINIMUM_DIFFICULTY uint64 = 10

	// Time window in blocks to calculate block time median
	MEDIAN_TIME_WINDOW uint64 = 11

	// Number of blocks used to calculate difficulty adjustments
	DIFFICULTY_ADJUST_WINDOW uint64 = 23

	// Average time span of the difficulty adjustment window
	BLOCK_TIME_WINDOW uint64 = DIFFICULTY_ADJUST_WINDOW * BLOCK_TIME_SEC

	// Maximum size time window used for difficulty adjustments
	UPPER_TIME_BOUND uint64 = BLOCK_TIME_WINDOW * 4 / 3

	// Minimum size time window used for difficulty adjustments
	LOWER_TIME_BOUND uint64 = BLOCK_TIME_WINDOW * 5 / 6
)
