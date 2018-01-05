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
	header := new(msgHeader)

	for {
		if err := header.Read(input); err != nil {
			return err
		}
		logrus.Debug("received header: ", header)

		if header.msgLen > maxMessageSize {
			return errors.New("too big message size")
		}

		// limit read
		rl := io.LimitReader(input, int64(header.msgLen))

		switch header.msgType {
		case msgTypePing:
			// update peer info & send pong
			var msg ping
			if err := msg.Read(rl); err != nil {
				return err
			}

			// update info
			p.Info.TotalDifficulty = msg.TotalDifficulty
			p.Info.Height = msg.Height

			// send pong
			// TODO: send actual blockchain state
			var resp pong
			resp.TotalDifficulty = consensus.Difficulty(1)
			resp.Height = 1
			WriteMessage(p.conn, resp)

		case msgTypePong:
			// update peer info
			var msg ping
			if err := msg.Read(rl); err != nil {
				return err
			}

			// update info
			p.Info.TotalDifficulty = msg.TotalDifficulty
			p.Info.Height = msg.Height
		case msgTypeGetPeerAddrs:
		case msgTypePeerAddrs:
		case msgTypeGetHeaders:
		case msgTypeHeaders:
		case msgTypeGetBlock:
		case msgTypeBlock:
		case msgTypeTransaction:

		default:
			return errors.New("receive unexpected message (type) from peer")
		}
	}
}

func (p Peer) Close() error {
	return p.conn.Close()
}