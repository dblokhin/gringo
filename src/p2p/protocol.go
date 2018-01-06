package p2p

import (

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

// Protocol defines grin-node network communicates
type Protocol interface {
	// TransmittedBytes bytes sent and received
	// TransmittedBytes() uint64

	// SendPing sends a Ping message to the remote peer. Will panic if handle has never
	// been called on this protocol.
	SendPing()

	// SendBlock sends a block to our remote peer
	SendBlock()
	SendTransaction()
	SendHeaderRequest()
	SendBlockRequest()
	SendPeerRequest()

	// Close the connection to the remote peer
	Close()
}
