// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package consensus

import (
)

// RangeProof of work
type Proof struct  {
	// The nonces
 	Nonces []uint32

	// The proof size
 	ProofSize uint
}

func NewProof(nonces []uint32) Proof {
	return Proof {
		Nonces: nonces,
		ProofSize: uint(len(nonces)), // TODO: it should be == ProofSize ?
	}
}