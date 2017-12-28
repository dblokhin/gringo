package p2p

import (
	"io"
	"bytes"
	"encoding/binary"
)

// Magic code expected in the header of every message
var magicCode = [2]uint8{0x1e, 0xc5}

const (
	// protocolVersion version of grin p2p protocol
	protocolVersion uint32 = 1

	// Size in bytes of a message header
	headerLen uint64 = 11

	// userAgent is name of version of the software
	userAgent = "gringo v0.0.1"
)

// messageType type of p2p protocol message
type messageType int

// Types of p2p messages
const (
	MsgError        messageType = iota
	MsgHand
	MsgShake
	MsgPing
	MsgPong
	MsgGetPeerAddrs
	MsgPeerAddrs
	MsgGetHeaders
	MsgHeaders
	MsgGetBlock
	MsgBlock
	MsgTransaction
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
	magic [2]uint8
	// msgType typo of the message.
	msgType messageType
	// msgLen length of the message in bytes.
	msgLen uint64
}

// msgBody is body of any protocol message
type msgBody interface {
	Len() uint64
	Write(writer io.Writer) error
}

func sendMessage(conn io.Writer, mType messageType, body msgBody) error {
	header := msgHeader{
		magic:   magicCode,
		msgType: mType,
		msgLen:  body.Len(),
	}

	if err := binary.Write(conn, binary.BigEndian, header); err != nil {
		return err
	}

	return body.Write(conn)
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