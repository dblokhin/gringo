// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package cuckoo

import (
	"golang.org/x/crypto/blake2b"
	"encoding/binary"
)

// New returns Cuckoo instance
func New(key []byte, sizeShift uint32) *Cuckoo {
	bsum := blake2b.Sum256(key)
	key = bsum[:]
	size := uint64(1) << sizeShift

	k0 := binary.LittleEndian.Uint64(key[:8])
	k1 := binary.LittleEndian.Uint64(key[7:16])

	v := make([]uint64, 4)
	v[0] = k0 ^ 0x736f6d6570736575
	v[1] = k1 ^ 0x646f72616e646f6d
	v[2] = k0 ^ 0x6c7967656e657261
	v[3] = k1 ^ 0x7465646279746573

	mask := (uint64(1) << sizeShift) / 2 - 1
	return &Cuckoo{
		mask,
		size,
		v,
	}
}

// Edge from u to v
type Edge struct {
	U uint64
	V uint64
}

// Cuckoo cycle context
type Cuckoo struct {
	mask uint64
	size uint64

	v []uint64
}

func (c *Cuckoo) newNode(nonce uint64, idx uint64) uint64 {
	return ((siphash24(c.v, 2 * nonce + idx) & c.mask) << 1) + idx
}

func (c *Cuckoo) NewEdge(nonce uint64) *Edge {
	return &Edge{
		U: c.newNode(nonce, 0),
		V: c.newNode(nonce, 1),
	}
}

func (c *Cuckoo) Verify(nonces []uint32, ease uint64) bool {
	return false
}
