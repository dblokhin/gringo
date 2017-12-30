package p2p

import (
	"net"
	"consensus"
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
	shake, err := handshake(conn)
	if err != nil {
		return nil, err
	}

	p := new(Peer)
	p.conn = conn
	p.Info.Version = shake.Version
	p.Info.Capabilities = shake.Capabilities
	p.Info.TotalDifficulty = shake.TotalDifficulty
	p.Info.UserAgent = shake.UserAgent

	return p, nil
}

func (p Peer) Write(b []byte) (n int, err error) {
	return p.conn.Write(b)
}

func (p Peer) Read(b []byte) (n int, err error) {
	return p.conn.Read(b)
}


