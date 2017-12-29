package p2p

import (
	"net"
	"consensus"
	"errors"
	"github.com/sirupsen/logrus"
)

// Peer is a participant of p2p network
type Peer struct {
	conn net.Conn
	hand hand

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

	logrus.Info("start new peer")
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	logrus.Info("peer connected")
	hand, err := handshake(conn)
	if err != nil {
		return nil, err
	}

	if hand.Version != protocolVersion {
		return nil, errors.New("incompatibility protocol version")
	}

	p := new(Peer)
	p.conn = conn
	p.Info.Version = hand.Version
	p.Info.Capabilities = hand.Capabilities
	p.Info.TotalDifficulty = hand.TotalDifficulty
	p.Info.Addr = hand.SenderAddr
	p.Info.UserAgent = hand.UserAgent

	return p, nil
}

/*func (p Peer) Write(b []byte) (n int, err error) {
	return p.conn.Write(b)
}*/


