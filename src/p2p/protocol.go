// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package p2p

import (
	"bufio"
	"errors"
	"github.com/dblokhin/gringo/src/consensus"
	"io"
)

const (
	// userAgent is name of version of the software
	userAgent = "gringo v0.0.1"
)

// Message defines methods for WriteMessage/ReadMessage functions
type Message interface {
	// Read reads from reader and fit self struct
	Read(r io.Reader) error

	// Bytes returns binary data of body message
	Bytes() []byte

	// Type says whats the message type should use in header
	Type() uint8
}

// WriteMessage writes to wr (net.conn) protocol message
func WriteMessage(w io.Writer, msg Message) (uint64, error) {
	data := msg.Bytes()

	header := Header{
		magic: consensus.MagicCode,
		Type:  msg.Type(),
		Len:   uint64(len(data)),
	}

	// use the buffered writer
	wr := bufio.NewWriter(w)
	if err := header.Write(wr); err != nil {
		return 0, err
	}

	n, err := wr.Write(data)
	if err != nil {
		return uint64(n) + consensus.HeaderLen, err
	}

	return uint64(n) + consensus.HeaderLen, wr.Flush()
}

// ReadMessage reads from r (net.conn) protocol message
func ReadMessage(r io.Reader, msg Message) (uint64, error) {
	var header Header

	// get the msg header
	rh := io.LimitReader(r, int64(consensus.HeaderLen))
	if err := header.Read(rh); err != nil {
		return 0, err
	}

	if header.Type != msg.Type() {
		return uint64(consensus.HeaderLen), errors.New("receive unexpected message type")
	}

	if header.Len > consensus.MaxMsgLen {
		return uint64(consensus.HeaderLen), errors.New("too big message size")
	}

	rb := io.LimitReader(r, int64(header.Len))
	return uint64(consensus.HeaderLen) + uint64(header.Len), msg.Read(rb)
}
