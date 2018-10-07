// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package cuckoo

import (
	"encoding/binary"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/blake2b"
)

// New returns a new instance of a Cuckoo cycle verifier. key is the data used
// to derived the siphash keys from and sizeShift is the number of vertices in
// the cuckoo graph as an exponent of 2.
func New(key []byte, sizeShift uint8) *Cuckoo {
	bsum := blake2b.Sum256(key)
	key = bsum[:]

	v := make([]uint64, 4)
	v[0] = binary.LittleEndian.Uint64(key[:8])
	v[1] = binary.LittleEndian.Uint64(key[8:16])
	v[2] = binary.LittleEndian.Uint64(key[16:24])
	v[3] = binary.LittleEndian.Uint64(key[24:32])

	numVertices := uint64(1) << sizeShift

	return &Cuckoo{
		mask: numVertices/2 - 1,
		size: numVertices,
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

	// size is the number of vertices in the cuckoo graph.
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
