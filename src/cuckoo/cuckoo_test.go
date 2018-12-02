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
