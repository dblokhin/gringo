package p2p

import (
	"io"
)

// Magic code expected in the header of every message
var magicCode = [2]byte{0x1e, 0xc5}

const (
	// protocolVersion version of grin p2p protocol
	protocolVersion uint32 = 1

	// size in bytes of a message header
	headerLen int64 = 11

	// userAgent is name of version of the software
	userAgent       = "gringo v0.0.1"
	maxStringLength = 1024 * 10
	// maxPeerAddresses in PeerAddrs resp
	maxPeerAddresses = 1024 * 10
	maxMessageSize  = 1024 * 1024 * 20	// TODO: check this for reality
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

func expectedResponse(requestType uint8) uint8 {
	switch requestType {
	case msgTypePing: return msgTypePong
	case msgTypeGetPeerAddrs: return msgTypePeerAddrs
	case msgTypeGetHeaders: return msgTypeHeaders
	case msgTypeGetBlock: return msgTypeBlock
	case msgTypeHand: return msgTypeShake
	}

	return msgTypeError
}

type Protocol interface {
	sendMsg(reader io.Reader)
	sendRequest(reader io.Reader)

	// transmittedBytes bytes sent and received
	transmittedBytes() uint64

	// sendPing sends a Ping message to the remote peer. Will panic if handle has never
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
