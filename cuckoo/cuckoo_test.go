// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package cuckoo

import (
	"encoding/binary"
	"encoding/hex"
	"testing"
)

func TestSum(t *testing.T) {
	if siphash24([4]uint64{1, 2, 3, 4}, 10) != uint64(928382149599306901) {
		t.Errorf("siphash24 was incorrect, want: %d.", uint64(928382149599306901))
	}
	if siphash24([4]uint64{1, 2, 3, 4}, 111) != uint64(10524991083049122233) {
		t.Errorf("siphash24 was incorrect, want: %d.", uint64(10524991083049122233))
	}
	if siphash24([4]uint64{9, 7, 6, 7}, 12) != uint64(1305683875471634734) {
		t.Errorf("siphash24 was incorrect, want: %d.", uint64(1305683875471634734))
	}
	if siphash24([4]uint64{9, 7, 6, 7}, 10) != uint64(11589833042187638814) {
		t.Errorf("siphash24 was incorrect, want: %d.", uint64(11589833042187638814))
	}
}

func TestBlock(t *testing.T) {
	if siphashBlock([4]uint64{1, 2, 3, 4}, 10) != uint64(1182162244994096396) {
		t.Errorf("siphashBlock was incorrect, want: %d.", uint64(1182162244994096396))
	}
	if siphashBlock([4]uint64{1, 2, 3, 4}, 123) != uint64(11303676240481718781) {
		t.Errorf("siphashBlock was incorrect, want: %d.", uint64(11303676240481718781))
	}
	if siphashBlock([4]uint64{9, 7, 6, 7}, 12) != uint64(4886136884237259030) {
		t.Errorf("siphashBlock was incorrect, want: %d.", uint64(4886136884237259030))
	}
}

func TestValidSolution(t *testing.T) {
	header := [80]byte{}

	// Replace the last four bytes of the key with the nonce.
	nonce := 20
	header[len(header)-4] = byte(nonce)
	header[len(header)-3] = byte(nonce << 8)
	header[len(header)-2] = byte(nonce << 16)
	header[len(header)-1] = byte(nonce << 24)

	cuckoo := NewCuckatoo(header[:], 29)

	k0, _ := hex.DecodeString("27580576fe290177")
	k1, _ := hex.DecodeString("f9ea9b2031f4e76e")
	k2, _ := hex.DecodeString("1663308c8607868f")
	k3, _ := hex.DecodeString("b88839b0fa180d0e")

	if binary.BigEndian.Uint64(k0) != cuckoo.v[0] {
		t.Errorf("Key derivation failed, got %x expected %x", cuckoo.v[0],
			binary.BigEndian.Uint64(k0))
	}
	if binary.BigEndian.Uint64(k1) != cuckoo.v[1] {
		t.Errorf("Key derivation failed, got %x expected %x", cuckoo.v[1],
			binary.BigEndian.Uint64(k1))
	}
	if binary.BigEndian.Uint64(k2) != cuckoo.v[2] {
		t.Errorf("Key derivation failed, got %x expected %x", cuckoo.v[2],
			binary.BigEndian.Uint64(k2))
	}
	if binary.BigEndian.Uint64(k3) != cuckoo.v[3] {
		t.Errorf("Key derivation failed, got %x expected %x", cuckoo.v[3],
			binary.BigEndian.Uint64(k3))
	}

	var V1 = []uint32{
		0x48a9e2, 0x9cf043, 0x155ca30, 0x18f4783, 0x248f86c, 0x2629a64, 0x5bad752, 0x72e3569,
		0x93db760, 0x97d3b37, 0x9e05670, 0xa315d5a, 0xa3571a1, 0xa48db46, 0xa7796b6, 0xac43611,
		0xb64912f, 0xbb6c71e, 0xbcc8be1, 0xc38a43a, 0xd4faa99, 0xe018a66, 0xe37e49c, 0xfa975fa,
		0x11786035, 0x1243b60a, 0x12892da0, 0x141b5453, 0x1483c3a0, 0x1505525e, 0x1607352c,
		0x16181fe3, 0x17e3a1da, 0x180b651e, 0x1899d678, 0x1931b0bb, 0x19606448, 0x1b041655,
		0x1b2c20ad, 0x1bd7a83c, 0x1c05d5b0, 0x1c0b9caa,
	}

	if !cuckoo.Verify(V1) {
		t.Error("Verify failed")
	}
}

func TestValidSolutionCuckaroo(t *testing.T) {
	key := [4]uint64{
		0x23796193872092ea,
		0xf1017d8a68c4b745,
		0xd312bd53d2cd307b,
		0x840acce5833ddc52,
	}
	expected := []uint32{
		0x45e9, 0x6a59, 0xf1ad, 0x10ef7, 0x129e8, 0x13e58, 0x17936, 0x19f7f, 0x208df, 0x23704,
		0x24564, 0x27e64, 0x2b828, 0x2bb41, 0x2ffc0, 0x304c5, 0x31f2a, 0x347de, 0x39686, 0x3ab6c,
		0x429ad, 0x45254, 0x49200, 0x4f8f8, 0x5697f, 0x57ad1, 0x5dd47, 0x607f8, 0x66199, 0x686c7,
		0x6d5f3, 0x6da7a, 0x6dbdf, 0x6f6bf, 0x6ffbb, 0x7580e, 0x78594, 0x785ac, 0x78b1d, 0x7b80d,
		0x7c11c, 0x7da35,
	}

	cuckoo := NewFromKeys(key)
	if !cuckoo.Verify(expected, 19) {
		t.Error("Verify failed")
	}
}

func TestShouldFindCycle(t *testing.T) {
	// Construct the example graph in figure 1 of the cuckoo cycle paper. The
	// cycle is: 8 -> 9 -> 4 -> 13 -> 10 -> 5 -> 8.

	edges := make([]*Edge, 6)
	edges[0] = &Edge{U: 8, V: 5}
	edges[1] = &Edge{U: 10, V: 5}
	edges[2] = &Edge{U: 4, V: 9}
	edges[3] = &Edge{U: 4, V: 13}
	edges[4] = &Edge{U: 8, V: 9}
	edges[5] = &Edge{U: 10, V: 13}

	if findCycleLength(edges) != 6 {
		t.Error("Verify failed")
	}
}

func TestShouldNotFindCycle(t *testing.T) {
	// Construct a path that isn't closed
	// 2 -> 5 -> 4 -> 9 -> 8 -> 11 -> 10

	edges := make([]*Edge, 6)
	edges[0] = &Edge{U: 1, V: 5}
	edges[1] = &Edge{U: 5, V: 4}
	edges[2] = &Edge{U: 4, V: 9}
	edges[3] = &Edge{U: 9, V: 8}
	edges[4] = &Edge{U: 8, V: 11}
	edges[5] = &Edge{U: 11, V: 10}

	cycle := findCycleLength(edges)
	if cycle != 0 {
		t.Errorf("Verify failed, found unexpected %d-cycle", cycle)
	}
}

func TestShouldNotFindCycleNotBipartite(t *testing.T) {
	// Construct a length 3 cycle that implies a non-bipartite graph.
	// 2 -> 4 -> 5 -> 2

	edges := make([]*Edge, 3)
	edges[0] = &Edge{U: 2, V: 4}
	edges[1] = &Edge{U: 4, V: 5}
	edges[2] = &Edge{U: 5, V: 2}

	cycle := findCycleLength(edges)
	if cycle != 0 {
		t.Errorf("Verify failed, found unexpected %d-cycle", cycle)
	}
}
