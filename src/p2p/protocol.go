package p2p

import (
	"io"
	"encoding/binary"
)

// Magic code expected in the header of every message
var magicCode = [2]byte{0x1e, 0xc5}

const (
	// protocolVersion version of grin p2p protocol
	protocolVersion uint32 = 1

	// Size in bytes of a message header
	headerLen int64 = 11

	// userAgent is name of version of the software
	userAgent = "gringo v0.0.1"
)

// Types of p2p messages
const (
	msgTypeError        uint8 = iota
	msgTypeHand
	msgTypeShake
	msgTypePing
	msgTypePong
	msgTypeGetPeerAddrs
	msgTypePeerAddrs
	msgTypeGetHeaders
	msgTypeHeaders
	msgTypeGetBlock
	msgTypeBlock
	msgTypeTransaction
)

// errCodes type
type errCodes int

// Error codes
const (
	unsupportedVersion errCodes = 100
)

type capabilities uint32

const (
	// We don't know (yet) what the peer can do.
	unknown capabilities = 0
	// Full archival node, has the whole history without any pruning.
	fullHist = 1 << 0
	// Can provide block headers and the UTXO set for some recent-enough height.
	utxoHist = 1 << 1
	// Can provide a list of healthy peers
	peerList = 1 << 2
	fullNode = fullHist | utxoHist | peerList
)

// msgHeader is header of any protocol message, used to identify incoming messages
type msgHeader struct {
	// magic number
	magic [2]byte
	// msgType typo of the message.
	msgType uint8
	// msgLen length of the message in bytes.
	msgLen uint64
}

func (h *msgHeader) Write(wr io.Writer) error {
	if _, err := wr.Write(h.magic[:]); err != nil {
		return err
	}
	if err := binary.Write(wr, binary.BigEndian, h.msgType); err != nil {
		return err
	}

	return binary.Write(wr, binary.BigEndian, h.msgLen)
}

func (h *msgHeader) Read(r io.Reader) error {
	if _, err := io.ReadFull(r, h.magic[:]); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &h.msgType); err != nil {
		return err
	}

	return binary.Read(r, binary.BigEndian, &h.msgLen)
}

/*
msgError        uint32 = iota
	msgHand
	msgShake
	msgPing
	msgPong
	msgGetPeerAddrs
	msgPeerAddrs
	msgGetHeaders
	msgHeaders
	msgGetBlock
	msgBlock
	msgTransaction

*/

type structReader interface {
	Read(r io.Reader) error
}

type Protocol interface {
	sendMsg(reader io.Reader)
	sendRequest(reader io.Reader)

	// transmittedBytes bytes sent and received
	transmittedBytes() uint64

	// sendPing sends a ping message to the remote peer. Will panic if handle has never
	// been called on this protocol.
	sendPing()

	// sendBlock sends a block to our remote peer
	sendBlock()
	sendTransaction()
	sendHeaderRequest()
	sendBlockRequest()
	sendPeerRequest()

	// close the connection to the remote peer
	close()
}
