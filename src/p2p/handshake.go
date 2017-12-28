package p2p

import (
	"net"
	"consensus"
)

// handshake by sender advertises its version and characteristics.
type handshake struct {
	// protocol version of the sender
	Version         uint32
	// capabilities of the sender
	Capabilities    capabilities
	// randomly generated for each handshake, helps detect self
	Nonce           uint64
	// total difficulty accumulated by the sender, used to check whether sync
	// may be needed
	TotalDifficulty consensus.Difficulty
	// network address of the sender
	SenderAddr      uint32	// SockAddr
	ReceiverAddr    uint32

	// name of version of the software
	UserAgent       string
}

func hand(conn net.Conn) (handshake, error){
	sendMessage(conn, MsgHand, )
}