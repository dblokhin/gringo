// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package consensus

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"github.com/dchest/siphash"
)

const (
	// The size of a short id used to identify inputs|outputs|kernels (6 bytes)
	ShortIDSize = 6
)

// Hash is hashes (block hash, commitments and so on)
type Hash []byte

// ShortID returns shortID from Hash
func (h Hash) ShortID(blockHash Hash) ShortID {
	result := make(ShortID, ShortIDSize+2)

	k0 := binary.LittleEndian.Uint64(blockHash[:8])
	k1 := binary.LittleEndian.Uint64(blockHash[8:16])

	hash := siphash.Hash(k0, k1, h)
	binary.LittleEndian.PutUint64(result, hash)

	// returned size is ShortIDSize
	return result[0:6]
}

type ShortID []byte

// String returns string representation
func (id ShortID) String() string {
	return hex.EncodeToString(id)
}

// ShortIDList sortable list of shortID
type ShortIDList []ShortID

func (s ShortIDList) Len() int {
	return len(s)
}

func (s ShortIDList) Less(i, j int) bool {
	return bytes.Compare(s[i], s[j]) < 0
}

func (s ShortIDList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
