// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package consensus

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestHash_ShortID(t *testing.T) {
	var hash Hash
	var expected ShortID
	otherHash := make(Hash, BlockHashSize)

	hash, _ = hex.DecodeString("81e47a19e6b29b0a65b9591762ce5143ed30d0261e5d24a3201752506b20f15c")
	expected, _ = hex.DecodeString("e973960ba690")

	if bytes.Compare(hash.ShortID(otherHash), expected) != 0 {
		t.Errorf("ShortID was incorrect, want: %s", expected.String())
	}

	hash, _ = hex.DecodeString("3a42e66e46dd7633b57d1f921780a1ac715e6b93c19ee52ab714178eb3a9f673")
	expected, _ = hex.DecodeString("f0c06e838e59")

	if bytes.Compare(hash.ShortID(otherHash), expected) != 0 {
		t.Errorf("ShortID was incorrect, want: %s", expected.String())
	}

	hash, _ = hex.DecodeString("3a42e66e46dd7633b57d1f921780a1ac715e6b93c19ee52ab714178eb3a9f673")
	expected, _ = hex.DecodeString("95bf0ca12d5b")
	otherHash, _ = hex.DecodeString("81e47a19e6b29b0a65b9591762ce5143ed30d0261e5d24a3201752506b20f15c")

	if bytes.Compare(hash.ShortID(otherHash), expected) != 0 {
		t.Errorf("ShortID was incorrect, got: %s", hash.ShortID(otherHash).String())
	}
}
