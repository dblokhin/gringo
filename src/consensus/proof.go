// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package consensus

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/dblokhin/gringo/src/cuckoo"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/blake2b"
)

// RangeProof of work
type Proof struct {
	// Power of 2 used for the size of the cuckoo graph
	EdgeBits uint8

	// The nonces
	Nonces []uint32
}

var (
	errInvalidPow = errors.New("invalid pow verify")
)

// Validate validates the pow
func (p *Proof) Validate(header *BlockHeader, cuckooSize uint8) error {
	logrus.Infof("block POW validate for size %d", cuckooSize)

	cuckoo := cuckoo.New(header.bytesWithoutPOW(), cuckooSize)
	if cuckoo.Verify(header.POW.Nonces, Easiness) {
		return nil
	}

	return errInvalidPow
}

// ToDifficulty converts the proof to a proof-of-work Target so they can be compared.
// Hashes the Cuckoo Proof data.
func (p *Proof) ToDifficulty() Difficulty {
	return MinimumDifficulty.FromHash(p.Hash())
}

// Hash returns hash of content pow
func (p *Proof) Hash() []byte {
	hash := blake2b.Sum256(p.Bytes())
	return hash[:]
}

// ProofBytes returns the serialised proof of work nonces.
func (p *Proof) ProofBytes() []byte {
	buff := new(bytes.Buffer)

	// The solution we serialise depends on the size of the cuckoo graph. The
	// cycle is always of length 42, but each vertex takes up more bits on
	// larger graphs, nonceLengthBits is this number of bits.
	nonceLengthBits := uint(p.EdgeBits)

	// Make a slice just large enough to fit all of the POW bits.
	bitvecLengthBits := nonceLengthBits * uint(ProofSize)
	bitvec := make([]uint8, (bitvecLengthBits+7)/8)

	for n, nonce := range p.Nonces {
		// Pack this nonce into the bit stream.
		for bit := uint(0); bit < nonceLengthBits; bit++ {
			// If this bit is set, then write it to the correct position in the
			// stream.
			if nonce&(1<<bit) != 0 {
				offsetBits := uint(n)*nonceLengthBits + bit
				bitvec[offsetBits/8] |= 1 << (offsetBits % 8)
			}
		}
	}

	if _, err := buff.Write(bitvec); err != nil {
		logrus.Fatal(err)
	}

	return buff.Bytes()
}

// Bytes returns binary []byte
func (p *Proof) Bytes() []byte {
	buff := new(bytes.Buffer)

	// Write size of cuckoo graph.
	if err := binary.Write(buff, binary.BigEndian, p.EdgeBits); err != nil {
		logrus.Fatal(err)
	}

	buff.Write(p.ProofBytes())

	return buff.Bytes()
}

func NewProof(nonces []uint32) Proof {
	return Proof{
		Nonces: nonces,
	}
}
