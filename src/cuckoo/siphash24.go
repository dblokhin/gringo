// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package cuckoo

const (
	siphashBlockBits = uint64(6)
	siphashBlockSize = uint64(1 << siphashBlockBits)
	siphashBlockMask = uint64(siphashBlockSize - 1)
)

// SipHash24 is an implementation of the siphash 2-4 keyed hash function by
// Jean-Philippe Aumasson and Daniel J. Bernstein.
type SipHash24 struct {
	// v is the current internal state.
	v [4]uint64
}

// NewSipHash24 returns a new instance of a SipHash24 hasher with the key v.
func NewSipHash24(v [4]uint64) SipHash24 {
	return SipHash24{
		v: v,
	}
}

// Sum64 outputs a 64-bit hash.
func (h *SipHash24) Sum64() uint64 {
	return h.v[0] ^ h.v[1] ^ h.v[2] ^ h.v[3]
}

// Write64 computes two, then four rounds of hashing.
func (h *SipHash24) Write64(nonce uint64) {
	h.v[3] ^= nonce

	round := func() {
		h.v[0] += h.v[1]
		h.v[1] = h.v[1]<<13 | h.v[1]>>(64-13)
		h.v[1] ^= h.v[0]
		h.v[0] = h.v[0]<<32 | h.v[0]>>(64-32)

		h.v[2] += h.v[3]
		h.v[3] = h.v[3]<<16 | h.v[3]>>(64-16)
		h.v[3] ^= h.v[2]

		h.v[0] += h.v[3]
		h.v[3] = h.v[3]<<21 | h.v[3]>>(64-21)
		h.v[3] ^= h.v[0]

		h.v[2] += h.v[1]
		h.v[1] = h.v[1]<<17 | h.v[1]>>(64-17)
		h.v[1] ^= h.v[2]
		h.v[2] = h.v[2]<<32 | h.v[2]>>(64-32)
	}

	round()
	round()

	h.v[0] ^= nonce
	h.v[2] ^= 0xff

	round()
	round()
	round()
	round()
}

// siphash24 computes a single siphash digest using the key v and a nonce.
func siphash24(v [4]uint64, nonce uint64) uint64 {
	h := NewSipHash24(v)
	h.Write64(nonce)
	return h.Sum64()
}

// siphashBlock computes a block of hashes of size siphashBlockSize.
func siphashBlock(v [4]uint64, nonce uint64) uint64 {
	siphash := NewSipHash24(v)

	// Find the start of the block that contains nonce.
	start := nonce &^ siphashBlockMask

	// Repeatedly hash from the start of the block to the end.
	var nonceHash uint64
	for n := start; n < start+siphashBlockSize; n++ {
		siphash.Write64(n)
		if n == nonce {
			nonceHash = siphash.Sum64()
		}
	}

	// Ensure the whole block of hashes has actually been calculated by xor-ing
	// the final state with nonceHash.
	if nonce == start+siphashBlockMask {
		return siphash.Sum64()
	} else {
		return nonceHash ^ siphash.Sum64()
	}
}
