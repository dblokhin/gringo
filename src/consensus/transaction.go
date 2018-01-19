// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package consensus

import (
	"fmt"
	"bytes"
	"github.com/sirupsen/logrus"
	"encoding/binary"
	"io"
	"secp256k1zkp"
	"errors"
)

// Transaction an grin transaction
type Transaction struct {
	// Set of inputs spent by the transaction
	Inputs []Input
	// Set of outputs the transaction produces
	Outputs []Output
	// Fee paid by the transaction
	Fee uint64
	// Transaction is not valid before this block height
	// It is invalid for this to be less than the lock_height of any UTXO being spent
	LockHeight uint64
	// The signature proving the excess is a valid public key, which signs
	// the transaction fee
	ExcessSig []byte

}

// Bytes implements p2p Message interface
func (t *Transaction) Bytes() []byte {
	buff := new(bytes.Buffer)

	// Write fee & lockHeight
	if err := binary.Write(buff, binary.BigEndian, t.Fee); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, t.LockHeight); err != nil {
		logrus.Fatal(err)
	}

	// Write ExcessSig
	if len(t.ExcessSig) > secp256k1zkp.MaxSignatureSize {
		logrus.Fatal(errors.New("invalid excess_sig len"))
	}
	if err := binary.Write(buff, binary.BigEndian, uint64(len(t.ExcessSig))); err != nil {
		logrus.Fatal(err)
	}

	if _, err := buff.Write(t.ExcessSig); err != nil {
		logrus.Fatal(err)
	}

	// Inputs & outputs lens
	if err := binary.Write(buff, binary.BigEndian, uint64(len(t.Inputs))); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, uint64(len(t.Outputs))); err != nil {
		logrus.Fatal(err)
	}

	// Consensus rule that everything is sorted in lexicographical order on the wire
	// FIXME: Write sorted input, output!
	// FIXME: Check sorted input, output, kernels requiring consensus rule!
	// Write inputs
	for _, input := range t.Inputs {
		if _, err := buff.Write(input.Commit); err != nil {
			logrus.Fatal(err)
		}
	}

	// Write outputs
	for _, output := range t.Outputs {
		if _, err := buff.Write(output.Bytes()); err != nil {
			logrus.Fatal(err)
		}
	}

	return buff.Bytes()
}

// Type implements p2p Message interface
func (t *Transaction) Type() uint8 {
	return MsgTypeTransaction
}

// Read implements p2p Message interface
func (t *Transaction) Read(r io.Reader) error {

	// Read fee & lockHeight
	if err := binary.Read(r, binary.BigEndian, &t.Fee); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Read(r, binary.BigEndian, &t.LockHeight); err != nil {
		logrus.Fatal(err)
	}

	// Read ExcessSig
	var excessSigLen uint64
	if err := binary.Read(r, binary.BigEndian, &excessSigLen); err != nil {
		return err
	}

	if excessSigLen > uint64(secp256k1zkp.MaxSignatureSize) {
		return errors.New("invalid excess_sig len")
	}

	t.ExcessSig = make([]byte, excessSigLen)
	if _, err := io.ReadFull(r, t.ExcessSig); err != nil {
		return err
	}

	// Inputs & outputs lens
	// FIXME: Check sorted input, output requiring consensus rule!
	var inputs, outputs uint64
	if err := binary.Read(r, binary.BigEndian, &inputs); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &outputs); err != nil {
		return err
	}

	t.Inputs = make([]Input, inputs)
	for i := uint64(0); i < inputs; i++ {
		commitment := make([]byte, secp256k1zkp.PedersenCommitmentSize)
		if _, err := io.ReadFull(r, commitment); err != nil {
			return err
		}

		t.Inputs[i].Commit = commitment
	}

	t.Outputs = make([]Output, outputs)
	for i := uint64(0); i < outputs; i++ {
		if err := t.Outputs[i].Read(r); err != nil {
			return err
		}
	}

	return nil
}

// String implements String() interface
func (t Transaction) String() string {
	return fmt.Sprintf("%#v", t)
}
