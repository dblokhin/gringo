// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package secp256k1zkp

import (
	"io"
	"fmt"
)

type Commitment []byte

// Bytes implements p2p Message interface
func (c *Commitment) Bytes() []byte {
	return *c
}

// Read implements p2p Message interface
func (c *Commitment) Read(r io.Reader) error {
	_, err := io.ReadFull(r, *c)

	return err
}

// String implements String() interface
func (p Commitment) String() string {
	return fmt.Sprintf("%#v", p)
}

type RangeProof struct {
	// The proof itself, at most 5134 bytes long
	Proof []byte // max size MAX_PROOF_SIZE
	// The length of the proof
	ProofLen int
}