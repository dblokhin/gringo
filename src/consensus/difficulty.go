// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package consensus

import (
	"time"
	"sort"
	"encoding/binary"
)

const (
	ZeroDifficulty  Difficulty = 0

	// The minimum mining difficulty we'll allow
	MinimumDifficulty Difficulty = 1
)

// Difficulty is defined as the maximum target divided by the block hash.
type Difficulty uint64

func (d Difficulty) FromNum(num uint64) Difficulty {
	return Difficulty(num)
}
// FromHash computes the difficulty from a hash. Divides the maximum target by the
// provided hash.
func (d Difficulty) FromHash(hash Hash) Difficulty {
	maxTarget := binary.BigEndian.Uint64(MAXTarget)

	// Use the first 64 bits of the given hash
	num := binary.BigEndian.Uint64(hash[:8])

	return Difficulty(maxTarget / num)
}

func (d Difficulty) IntoNum() uint64 {
	return uint64(d)
}

// NextDifficulty computes the proof-of-work difficulty that the next block should comply
// with. Takes an iterator over past blocks, from latest (highest height) to
// oldest (lowest height). The iterator produces pairs of timestamp and
// difficulty for each block.
//
// The difficulty calculation is based on both Digishield and GravityWave
// family of difficulty computation, coming to something very close to Zcash.
// The refence difficulty is an average of the difficulty over a window of
// DIFFICULTY_ADJUST_WINDOW blocks. The corresponding timespan is calculated by using the
// difference between the median timestamps at the beginning and the end
// of the window.

func NextDifficulty(blist BlockList) Difficulty {

	blen := len(blist)
	if blen == 0 {
		return ZeroDifficulty
	}

	// Sum of difficulties in the window, used to calculate the average later.
	sumDiff := ZeroDifficulty

	// Block times at the begining and end of the adjustment window, used to
	// calculate medians later.
	windowBegin := make([]time.Time, 0)
	windowEnd := make([]time.Time, 0)

	for i := blen - 1; i >= 0; i-- {
		if i < DifficultyAdjustWindow {
			sumDiff += blist[i].Header.Difficulty

			if i < MedianTimeWindow {
				windowBegin = append(windowBegin, blist[i].Header.Timestamp)
			}
		} else {
			if i < DifficultyAdjustWindow + MedianTimeWindow {
				windowEnd = append(windowEnd, blist[i].Header.Timestamp)
			} else {
				break
			}
		}
	}

	// Check we have enough blocks
	if len(windowEnd) < MedianTimeWindow {
		return MinimumDifficulty
	}

	// Calculating time medians at the beginning and end of the window.
	sort.SliceStable(windowBegin, func(i, j int) bool {
		return windowBegin[i].Before(windowBegin[j])
	})
	sort.SliceStable(windowEnd, func(i, j int) bool {
		return windowEnd[i].Before(windowEnd[j])
	})

	beginTime := windowBegin[len(windowBegin) / 2]
	endTime := windowEnd[len(windowEnd) / 2]

	// Average difficulty and dampened average time
	diffAvg := sumDiff / MinimumDifficulty.FromNum(uint64(DifficultyAdjustWindow))
	ts := (3 * BlockTimeWindow + beginTime.Sub(endTime)) / 4

	// Apply time bounds
	if ts < LowerTimeBound {
		ts = LowerTimeBound
	}
	if ts > UpperTimeBound {
		ts = UpperTimeBound
	}

	//Result
	diff := diffAvg * MinimumDifficulty.FromNum(uint64(BlockTimeWindow)) / MinimumDifficulty.FromNum(uint64(ts))
	if diff > MinimumDifficulty {
		return diff
	}

	return MinimumDifficulty
}