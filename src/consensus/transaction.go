// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package consensus

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"sort"
)

// Transaction an grin transaction
type Transaction struct {
	// The "k2" kernel offset.
	KernelOffset Hash
	// Set of inputs spent by the transaction
	Inputs InputList
	// Set of outputs the transaction produces
	Outputs OutputList
	// The kernels for this transaction
	Kernels TxKernelList
}

// Bytes implements p2p Message interface
func (t *Transaction) Bytes() []byte {
	buff := new(bytes.Buffer)

	if _, err := buff.Write(t.KernelOffset); err != nil {
		logrus.Fatal(err)
	}

	// Inputs & outputs lens
	if err := binary.Write(buff, binary.BigEndian, uint64(len(t.Inputs))); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, uint64(len(t.Outputs))); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, uint64(len(t.Kernels))); err != nil {
		logrus.Fatal(err)
	}

	// Consensus rule that everything is sorted in lexicographical order on the wire
	// consensus rule: input, output, kernels MUST BE sorted!
	sort.Sort(t.Inputs)
	sort.Sort(t.Outputs)
	sort.Sort(t.Kernels)

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

	// Write kernels
	for _, kernel := range t.Kernels {
		if _, err := buff.Write(kernel.Bytes()); err != nil {
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
	t.KernelOffset = make([]byte, 32)
	if _, err := io.ReadFull(r, t.KernelOffset); err != nil {
		return err
	}

	// Read the lengths of the subsequent fields.
	var inputs, outputs, kernels uint64
	if err := binary.Read(r, binary.BigEndian, &inputs); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &outputs); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &kernels); err != nil {
		return err
	}

	// Sanity check the lengths.
	if inputs > 1000000 {
		return errors.New("transaction contains too many inputs")
	}
	if outputs > 1000000 {
		return errors.New("transaction contains too many outputs")
	}
	if kernels > 1000000 {
		return errors.New("transaction contains too many kernels")
	}

	t.Inputs = make([]Input, inputs)
	for i := uint64(0); i < inputs; i++ {
		if err := t.Inputs[i].Read(r); err != nil {
			return err
		}
	}

	t.Outputs = make([]Output, outputs)
	for i := uint64(0); i < outputs; i++ {
		if err := t.Outputs[i].Read(r); err != nil {
			return err
		}
	}

	t.Kernels = make([]TxKernel, kernels)
	for i := uint64(0); i < kernels; i++ {
		if err := t.Kernels[i].Read(r); err != nil {
			return err
		}
	}

	// TODO: Check block weight.

	// Check sorted input, output requiring consensus rule!
	if !sort.IsSorted(t.Inputs) {
		return errors.New("consensus error: inputs are not sorted")
	}

	if !sort.IsSorted(t.Outputs) {
		return errors.New("consensus error: outputs are not sorted")
	}

	if !sort.IsSorted(t.Kernels) {
		return errors.New("consensus error: kernels are not sorted")
	}

	return nil
}

// String implements String() interface
func (t Transaction) String() string {
	return fmt.Sprintf("%#v", t)
}
