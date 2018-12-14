// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package consensus

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/sirupsen/logrus"
	"io"
)

type Locator struct {
	Hashes []Hash
}

// Bytes implements Message interface
func (h *Locator) Bytes() []byte {
	buff := new(bytes.Buffer)

	// check the bounds & set the limits
	if len(h.Hashes) > MaxLocators {
		logrus.Fatal(errors.New("invalid hashes len in locator"))
	}

	if err := binary.Write(buff, binary.BigEndian, uint8(len(h.Hashes))); err != nil {
		logrus.Fatal(err)
	}

	for _, hash := range h.Hashes {
		if _, err := buff.Write(hash); err != nil {
			logrus.Fatal(err)
		}
	}

	return buff.Bytes()
}

// Read implements Message interface
func (h *Locator) Read(r io.Reader) error {

	var count uint8
	if err := binary.Read(r, binary.BigEndian, &count); err != nil {
		return err
	}

	if int(count) > MaxLocators {
		return errors.New("too big locator len from peer")
	}

	h.Hashes = make([]Hash, count)
	for i := 0; i < int(count); i++ {
		h.Hashes[i] = make([]byte, BlockHashSize)

		if _, err := io.ReadFull(r, h.Hashes[i]); err != nil {
			return err
		}
	}

	return nil
}
