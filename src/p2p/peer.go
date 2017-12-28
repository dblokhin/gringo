package p2p

import (
	"net"
	"consensus"
)

// Peer is a participant of p2p network
type Peer struct {
	conn net.Conn
	hand handshake

	// Info connected peer
	Info struct{
		// protocol version of the sender
		Version         uint32
		// capabilities of the sender
		Capabilities    capabilities
		// total difficulty accumulated by the sender, used to check whether sync
		// may be needed
		TotalDifficulty consensus.Difficulty
		// network address of the sender
		Addr      		net.Addr	// SockAddr
		// name of version of the software
		UserAgent       string
	}

}

// NewPeer connects to peer
func NewPeer(addr string) (*Peer, error) {

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	hand, err := hand(conn)
	if err != nil {
		return nil, err
	}

	p := new(Peer)
	p.conn = conn
	p.hand = hand

	return p, nil
}

/*func (p Peer) Write(b []byte) (n int, err error) {
	return p.conn.Write(b)
}*/


