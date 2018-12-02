// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package cuckoo

import (
	"encoding/binary"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/blake2b"
)

// New returns Cuckoo instance
func New(key []byte, sizeShift uint32) *Cuckoo {
	bsum := blake2b.Sum256(key)
	key = bsum[:]

	k0 := binary.LittleEndian.Uint64(key[:8])
	k1 := binary.LittleEndian.Uint64(key[8:16])

	v := make([]uint64, 4)
	v[0] = k0 ^ 0x736f6d6570736575
	v[1] = k1 ^ 0x646f72616e646f6d
	v[2] = k0 ^ 0x6c7967656e657261
	v[3] = k1 ^ 0x7465646279746573

	return &Cuckoo{
		mask: (uint64(1)<<sizeShift)/2 - 1,
		size: uint64(1) << sizeShift,
		v:    v,
	}
}

// Edge from u to v
type Edge struct {
	U uint64
	V uint64

	usedU bool
	usedV bool
}

// Cuckoo cycle context
type Cuckoo struct {
	mask uint64
	size uint64

	v []uint64
}

func (c *Cuckoo) newNode(nonce uint64, i uint64) uint64 {
	return ((siphash24(c.v, 2*nonce+i) & c.mask) << 1) | i
}

func (c *Cuckoo) NewEdge(nonce uint32) *Edge {
	return &Edge{
		U: c.newNode(uint64(nonce), 0),
		V: c.newNode(uint64(nonce), 1),
	}
}

func (c *Cuckoo) Verify(nonces []uint32, ease uint64) bool {
	proofSize := len(nonces)

	// zero proof is always invalid
	if proofSize == 0 {
		return false
	}

	easiness := ease * c.size / 100

	// Preparing edges
	proof := make([]*Edge, proofSize)
	for i := 0; i < proofSize; i++ {
		if uint64(nonces[i]) >= easiness || (i != 0 && nonces[i] <= nonces[i-1]) {
			return false
		}

		proof[i] = c.NewEdge(nonces[i])
		logrus.Debugf("%#v", *proof[i])
	}

	// Checking edges
	i := 0    // first node
	flag := 0 // flag indicates what we need compare U or V
	cycle := 0

loop:
	for {
		if flag%2 == 0 {
			for j := 0; j < proofSize; j++ {
				if j != i && !proof[j].usedU && proof[i].U == proof[j].U {
					proof[i].usedU = true
					proof[j].usedU = true

					i = j
					flag ^= 1
					cycle++

					continue loop
				}
			}
		} else {
			for j := 0; j < proofSize; j++ {
				if j != i && !proof[j].usedV && proof[i].V == proof[j].V {
					proof[i].usedV = true
					proof[j].usedV = true

					i = j
					flag ^= 1
					cycle++

					continue loop
				}
			}
		}

		break
	}

	return cycle == proofSize
}
