// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package consensus

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"encoding/binary"
	"errors"
	"golang.org/x/crypto/blake2b"
	"cuckoo"
)

// RangeProof of work
type Proof struct  {
	// The nonces
 	Nonces []uint32

	// The proof size
 	ProofSize uint
}

// Validate validates the pow
func (p *Proof) Validate(bh *BlockHeader, cuckooSize uint32) error {
	logrus.Info("block POW validate")

	cuckoo := cuckoo.New(bh.Hash(), cuckooSize)
	if cuckoo.Verify(bh.POW.Nonces, Easiness) {
		return nil
	}

	return errors.New("invalid pow verify")
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

// Bytes returns binary []byte
func (p *Proof) Bytes() []byte {
	buff := new(bytes.Buffer)

	// Write POW
	if uint32(p.ProofSize) != ProofSize || uint32(len(p.Nonces)) != ProofSize {
		logrus.Fatal(errors.New("invalid proof len"))
	}

	for i := 0; i < int(ProofSize); i++ {
		if err := binary.Write(buff, binary.BigEndian, p.Nonces[i]); err != nil {
			logrus.Fatal(err)
		}
	}

	return buff.Bytes()
}

func NewProof(nonces []uint32) Proof {
	return Proof {
		Nonces: nonces,
		ProofSize: uint(len(nonces)), // TODO: it should be == ProofSize ?
	}
}