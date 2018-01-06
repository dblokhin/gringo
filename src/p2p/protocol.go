package p2p

import (

)

const (
	// userAgent is name of version of the software
	userAgent       = "gringo v0.0.1"
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
