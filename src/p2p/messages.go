// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package p2p

import (
	"io"
	"encoding/binary"
	"github.com/sirupsen/logrus"
	"consensus"
	"bytes"
	"net"
	"errors"
	"fmt"
)

// Header is header of any protocol message, used to identify incoming messages
type Header struct {
	// magic number
	magic [2]byte
	// Type typo of the message.
	Type uint8
	// Len length of the message in bytes.
	Len uint64
}

// Write writes header as binary data to writer
func (h *Header) Write(wr io.Writer) error {
	if _, err := wr.Write(h.magic[:]); err != nil {
		return err
	}
	if err := binary.Write(wr, binary.BigEndian, h.Type); err != nil {
		return err
	}

	return binary.Write(wr, binary.BigEndian, h.Len)
}

// Read reads from reader & fill struct
func (h *Header) Read(r io.Reader) error {
	if _, err := io.ReadFull(r, h.magic[:]); err != nil {
		return err
	}

	if !h.validateMagic() {
		logrus.Debug("got magic: ", h.magic[:])
		return errors.New("invalid magic code")
	}

	if err := binary.Read(r, binary.BigEndian, &h.Type); err != nil {
		return err
	}

	return binary.Read(r, binary.BigEndian, &h.Len)
}

// validateMagic verifies magic code
func (h Header) validateMagic() bool {
	return h.magic[0] == 0x1e && h.magic[1] == 0xc5
}

// Ping request
type Ping struct {
	// total difficulty accumulated by the sender, used to check whether sync
	// may be needed
	TotalDifficulty consensus.Difficulty
	// total height
	Height uint64
}

// Bytes implements Message interface
func (p *Ping) Bytes() []byte {
	buff := new(bytes.Buffer)

	if err := binary.Write(buff, binary.BigEndian, uint64(p.TotalDifficulty)); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, uint64(p.Height)); err != nil {
		logrus.Fatal(err)
	}

	return buff.Bytes()
}

// Type implements Message interface
func (p *Ping) Type() uint8 {
	return consensus.MsgTypePing
}

// Read implements Message interface
func (p *Ping) Read(r io.Reader) error {

	if err := binary.Read(r, binary.BigEndian, (*uint64)(&p.TotalDifficulty)); err != nil {
		return err
	}

	return binary.Read(r, binary.BigEndian, (*uint64)(&p.Height))
}

// String implements String() interface
func (p Ping) String() string {
	return fmt.Sprintf("%#v", p)
}

// Pong response same as Ping
type Pong struct {
	Ping
}

// Type implements Messagee interface
func (p *Pong) Type() uint8 {
	return consensus.MsgTypePong
}

// GetPeerAddrs asks for other peers addresses, required for network discovery.
type GetPeerAddrs struct {
	// filters on the capabilities we'd like the peers to have
	Capabilities consensus.Capabilities
}

// Bytes implements Message interface
func (p *GetPeerAddrs) Bytes() []byte {
	buff := new(bytes.Buffer)

	if err := binary.Write(buff, binary.BigEndian, uint32(p.Capabilities)); err != nil {
		logrus.Fatal(err)
	}

	return buff.Bytes()
}

// Type implements Message interface
func (p *GetPeerAddrs) Type() uint8 {
	return consensus.MsgTypeGetPeerAddrs
}

// Read implements Message interface
func (p *GetPeerAddrs) Read(r io.Reader) error {

	return binary.Read(r, binary.BigEndian, (*uint32)(&p.Capabilities))
}

// String implements String() interface
func (p GetPeerAddrs) String() string {
	return fmt.Sprintf("%#v", p)
}

// PeerError sending an error back (usually followed  by closing conn)
type PeerError struct {
	// error code
	Code uint32
	// slightly more user friendly message
	Message string
}

// Bytes implements Message interface
func (p *PeerError) Bytes() []byte {
	buff := new(bytes.Buffer)

	if err := binary.Write(buff, binary.BigEndian, uint32(p.Code)); err != nil {
		logrus.Fatal(err)
	}

	// Write user agent [len][string]
	if err := binary.Write(buff, binary.BigEndian, uint64(len(p.Message))); err != nil {
		logrus.Fatal(err)
	}
	buff.WriteString(p.Message)
	return buff.Bytes()
}

// Type implements Message interface
func (p *PeerError) Type() uint8 {
	return consensus.MsgTypeError
}

// Read implements Message interface
func (p *PeerError) Read(r io.Reader) error {

	if err := binary.Read(r, binary.BigEndian, (*uint32)(&p.Code)); err != nil {
		return err
	}

	var messageLen uint64
	if err := binary.Read(r, binary.BigEndian, &messageLen); err != nil {
		return err
	}

	buff := make([]byte, messageLen)
	if _, err := io.ReadFull(r, buff); err != nil {
		return err
	}

	p.Message = string(buff)
	return nil
}

// String implements String() interface
func (p PeerError) String() string {
	return fmt.Sprintf("%#v", p)
}

// PeerAddrs we know of that are fresh enough, in response to GetPeerAddrs
type PeerAddrs struct {
	peers []*net.TCPAddr
}

// Bytes implements Message interface
func (p *PeerAddrs) Bytes() []byte {
	buff := new(bytes.Buffer)

	if len(p.peers) > consensus.MaxPeerAddrs {
		logrus.Debug("length peers: ", len(p.peers))
		logrus.Fatal(errors.New("too big peer addrs count for sending"))
	}

	if err := binary.Write(buff, binary.BigEndian, uint32(len(p.peers))); err != nil {
		logrus.Fatal(err)
	}

	for _, peerAddr := range p.peers {
		serializeTCPAddr(buff, peerAddr)
	}

	return buff.Bytes()
}

// Type implements Message interface
func (p *PeerAddrs) Type() uint8 {
	return consensus.MsgTypePeerAddrs
}

// Read implements Message interface
func (p *PeerAddrs) Read(r io.Reader) error {
	var peersCount uint32

	if err := binary.Read(r, binary.BigEndian, &peersCount); err != nil {
		return err
	}

	if peersCount > consensus.MaxPeerAddrs {
		return errors.New("too big peer addrs count")
	}

	for i := uint32(0); i < peersCount; i++ {
		addr, err := deserializeTCPAddr(r)
		if err != nil {
			return err
		}

		p.peers = append(p.peers, addr)
	}

	return nil
}

// String implements String() interface
func (p PeerAddrs) String() string {
	return fmt.Sprintf("%#v", p)
}

// GetBlock message for requesting block by hash
type GetBlock struct {
	Hash consensus.BlockHash
}

// Bytes implements Message interface
func (h *GetBlock) Bytes() []byte {
	if len(h.Hash) != consensus.BlockHashSize {
		logrus.Fatal(errors.New("invalid block hash len"))
	}

	return h.Hash
}

// Type implements Message interface
func (h *GetBlock) Type() uint8 {
	return consensus.MsgTypeGetBlock
}

// Read implements Message interface
func (h *GetBlock) Read(r io.Reader) error {

	hash := make([]byte, consensus.BlockHashSize)
	_, err := io.ReadFull(r, hash)

	h.Hash = hash
	return err
}

// String implements String() interface
func (h GetBlock) String() string {
	return fmt.Sprintf("%#v", h)
}

// BlockHeaders message with grin headers
type BlockHeaders struct {
	Headers []consensus.BlockHeader
}

// Bytes implements Message interface
func (h *BlockHeaders) Bytes() []byte {
	buff := new(bytes.Buffer)

	// check the bounds of h.Headers & set the limits
	if len(h.Headers) > consensus.MaxBlockHeaders {
		logrus.Fatal(errors.New("invalid headers len in BlockHeaders"))
	}

	if err := binary.Write(buff, binary.BigEndian, uint16(len(h.Headers))); err != nil {
		logrus.Fatal(err)
	}

	for _, header := range h.Headers {
		if _, err := buff.Write(header.Bytes()); err != nil {
			logrus.Fatal(err)
		}
	}

	return buff.Bytes()
}

// Type implements Message interface
func (h *BlockHeaders) Type() uint8 {
	return consensus.MsgTypeHeaders
}

// Read implements Message interface
func (h *BlockHeaders) Read(r io.Reader) error {

	var count uint16
	if err := binary.Read(r, binary.BigEndian, &count); err != nil {
		return err
	}

	if int(count) > consensus.MaxBlockHeaders {
		return errors.New("too big block headers len from peer")
	}

	h.Headers = make([]consensus.BlockHeader, count)
	for i := 0; i < int(count); i++ {
		if err := h.Headers[i].Read(r); err != nil {
			return err
		}
	}

	return nil
}

// String implements String() interface
func (h BlockHeaders) String() string {
	return fmt.Sprintf("%#v", h)
}

// GetBlockHeaders message for requesting headers
type GetBlockHeaders struct {
	Locator consensus.Locator
}

// Bytes implements Message interface
func (h *GetBlockHeaders) Bytes() []byte {
	return h.Locator.Bytes()
}

// Type implements Message interface
func (h *GetBlockHeaders) Type() uint8 {
	return consensus.MsgTypeGetHeaders
}

// Read implements Message interface
func (h *GetBlockHeaders) Read(r io.Reader) error {
	return h.Locator.Read(r)
}

// String implements String() interface
func (h GetBlockHeaders) String() string {
	return fmt.Sprintf("%#v", h)
}
