package p2p

import (
	"net"
	"consensus"
	"github.com/sirupsen/logrus"
	"bufio"
	"io"
	"errors"
)

// Peer is a participant of p2p network
type Peer struct {
	conn net.Conn
	hand hand

	// Info connected peer
	Info struct {
		// protocol version of the sender
		Version uint32
		// capabilities of the sender
		Capabilities capabilities
		// total difficulty accumulated by the sender, used to check whether sync
		// may be needed
		TotalDifficulty consensus.Difficulty
		// name of version of the software
		UserAgent string
		// Height
		Height uint64
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

// HandleLoop starts event loop listening
func (p *Peer) HandleLoop() error {
	input := bufio.NewReader(p.conn)
	header := new(Header)

	for {
		if err := header.Read(input); err != nil {
			return err
		}
		logrus.Debug("received header: ", header)

		if header.Len > maxMessageSize {
			return errors.New("too big message size")
		}

		// limit read
		rl := io.LimitReader(input, int64(header.Len))

		switch header.Type {
		case msgTypePing:
			// update peer info & send Pong
			var msg Ping
			if err := msg.Read(rl); err != nil {
				return err
			}

			// update info
			p.Info.TotalDifficulty = msg.TotalDifficulty
			p.Info.Height = msg.Height

			logrus.Debug("received Ping: ", msg)
			// send Pong
			// TODO: send actual blockchain state
			var resp Pong
			resp.TotalDifficulty = consensus.Difficulty(1)
			resp.Height = 1

			if err := WriteMessage(p.conn, resp); err != nil {
				return err
			}

		case msgTypePong:
			// update peer info
			var msg Pong
			if err := msg.Read(rl); err != nil {
				return err
			}

			// update info
			p.Info.TotalDifficulty = msg.TotalDifficulty
			p.Info.Height = msg.Height

			logrus.Debug("received Pong: ", msg)

		case msgTypeGetPeerAddrs:
			var msg GetPeerAddrs
			if err := msg.Read(rl); err != nil {
				return err
			}
			logrus.Info("received msgTypeGetPeerAddrs")

			// Send answer
			var resp PeerAddrs
			if err := WriteMessage(p.conn, resp); err != nil {
				return err
			}

		case msgTypePeerAddrs:
			var msg PeerAddrs
			if err := msg.Read(rl); err != nil {
				return err
			}
			logrus.Info("received msgTypePeerAddrs")
		case msgTypeGetHeaders:
			logrus.Info("received msgTypeGetHeaders")
		case msgTypeHeaders:
			logrus.Info("received msgTypeHeaders")
		case msgTypeGetBlock:
			logrus.Info("received msgTypeGetBlock")
		case msgTypeBlock:
			logrus.Info("received msgTypeBlock")
		case msgTypeTransaction:
			logrus.Info("received msgTypeTransaction")

		default:
			return errors.New("receive unexpected message (type) from peer")
		}
	}
}

// Close peer
func (p Peer) Close() error {
	return p.conn.Close()
}

// SendPing sends Ping request to peer
func (p Peer) SendPing() error {
	var request Ping
	request.TotalDifficulty = consensus.Difficulty(1)
	request.Height = 1

	return WriteMessage(p.conn, request)
}