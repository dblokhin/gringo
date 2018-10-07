// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package cuckoo

import "testing"

func TestSum(t *testing.T) {
	if siphash24([]uint64{1, 2, 3, 4}, 10) != uint64(928382149599306901) {
		t.Errorf("siphash24 was incorrect, want: %d.", uint64(928382149599306901))
	}
	if siphash24([]uint64{1, 2, 3, 4}, 111) != uint64(10524991083049122233) {
		t.Errorf("siphash24 was incorrect, want: %d.", uint64(10524991083049122233))
	}
	if siphash24([]uint64{9, 7, 6, 7}, 12) != uint64(1305683875471634734) {
		t.Errorf("siphash24 was incorrect, want: %d.", uint64(1305683875471634734))
	}
	if siphash24([]uint64{9, 7, 6, 7}, 10) != uint64(11589833042187638814) {
		t.Errorf("siphash24 was incorrect, want: %d.", uint64(11589833042187638814))
	}
}

var V1 = []uint32{
	0x3bbd, 0x4e96, 0x1013b, 0x1172b, 0x1371b, 0x13e6a, 0x1aaa6, 0x1b575,
	0x1e237, 0x1ee88, 0x22f94, 0x24223, 0x25b4f, 0x2e9f3, 0x33b49, 0x34063,
	0x3454a, 0x3c081, 0x3d08e, 0x3d863, 0x4285a, 0x42f22, 0x43122, 0x4b853,
	0x4cd0c, 0x4f280, 0x557d5, 0x562cf, 0x58e59, 0x59a62, 0x5b568, 0x644b9,
	0x657e9, 0x66337, 0x6821c, 0x7866f, 0x7e14b, 0x7ec7c, 0x7eed7, 0x80643,
	0x8628c, 0x8949e,
}

func TestValidSolution(t *testing.T) {
	header := []byte{49}
	cuckoo := New(header, 20)
	if !cuckoo.Verify(V1, 75) {
		t.Error("Verify failed")
	}
}

func TestShouldFindCycle(t *testing.T) {
	header := []byte{49}
	cuckoo := New(header, 20)

	// Construct the example graph in figure 1 of the cuckoo cycle paper. The
	// cycle is: 8 -> 9 -> 4 -> 13 -> 10 -> 5 -> 8.

	edges := make([]*Edge, 6)
	edges[0] = &Edge{U: 8, V: 5}
	edges[1] = &Edge{U: 10, V: 5}
	edges[2] = &Edge{U: 4, V: 9}
	edges[3] = &Edge{U: 4, V: 13}
	edges[4] = &Edge{U: 8, V: 9}
	edges[5] = &Edge{U: 10, V: 13}

	if cuckoo.findCycleLength(edges) != 6 {
		t.Error("Verify failed")
	}
}

func TestShouldNotFindCycle(t *testing.T) {
	cuckoo := New([]byte{49}, 20)

	// Construct a path that isn't closed
	// 2 -> 5 -> 4 -> 9 -> 8 -> 11 -> 10

	edges := make([]*Edge, 6)
	edges[0] = &Edge{U: 1, V: 5}
	edges[1] = &Edge{U: 5, V: 4}
	edges[2] = &Edge{U: 4, V: 9}
	edges[3] = &Edge{U: 9, V: 8}
	edges[4] = &Edge{U: 8, V: 11}
	edges[5] = &Edge{U: 11, V: 10}

	cycle := cuckoo.findCycleLength(edges)
	if cycle != 0 {
		t.Errorf("Verify failed, found unexpected %d-cycle", cycle)
	}
}

func TestShouldNotFindCycleNotBipartite(t *testing.T) {
	cuckoo := New([]byte{49}, 20)

	// Construct a length 3 cycle that implies a non-bipartite graph.
	// 2 -> 4 -> 5 -> 2

	edges := make([]*Edge, 3)
	edges[0] = &Edge{U: 2, V: 4}
	edges[1] = &Edge{U: 4, V: 5}
	edges[2] = &Edge{U: 5, V: 2}

	cycle := cuckoo.findCycleLength(edges)
	if cycle != 0 {
		t.Errorf("Verify failed, found unexpected %d-cycle", cycle)
	}
}
