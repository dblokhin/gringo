package p2p

import (
	"net"
	"consensus"
	"github.com/sirupsen/logrus"
	"bufio"
	"io"
	"errors"
	"sync"
	"sync/atomic"
)

// Peer is a participant of p2p network
type Peer struct {
	conn net.Conn

	// The following fields are only meant to be used *atomically*
	bytesReceived uint64
	bytesSent     uint64

	quit      chan struct{}
	wg        sync.WaitGroup
	sendQueue chan SendableMessage

	hand hand

	// Info connected peer
	Info struct {
		// protocol version of the sender
		Version uint32
		// capabilities of the sender
		Capabilities consensus.Capabilities
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
	p.quit = make(chan struct{})
	p.sendQueue = make(chan SendableMessage)

	p.Info.Version = shake.Version
	p.Info.Capabilities = shake.Capabilities
	p.Info.TotalDifficulty = shake.TotalDifficulty
	p.Info.UserAgent = shake.UserAgent

	return p, nil
}

// Start starts loop listening, write handler and so on
func (p Peer) Start() {
	p.wg.Add(2)
	go p.writeHandler()
	go p.readHandler()
}

// writeHandler is a goroutine dedicated to reading messages off of an incoming
// queue, and writing them out to the wire.
//
// NOTE: This method MUST be run as a goroutine.
func (p Peer) writeHandler() {
	var exitError error

out:
	for {
		select {
		case msg := <-p.sendQueue:
			if exitError = WriteMessage(p.conn, msg); exitError != nil {
				break out
			}
			// FIXME: msg.Bytes() call is not eficient
			atomic.AddUint64(&p.bytesSent, uint64(len(msg.Bytes()))+consensus.HeaderLen)
		case <-p.quit:
			exitError = errors.New("peer exiting")
			break out
		}
	}

	p.wg.Done()
	p.Disconnect(exitError)
}

// queueMessage places msg to send queue
func (p Peer) queueMessage(msg SendableMessage) {
	select {
	case <-p.quit: logrus.Info("cannot send message, peer is shutting down")
	case p.sendQueue <- msg:
	}
}

// readHandler is responsible for reading messages off the wire in series, then
// properly dispatching the handling of the message to the proper subsystem.
//
// NOTE: This method MUST be run as a goroutine.
func (p Peer) readHandler() {
	var exitError error
	input := bufio.NewReader(p.conn)
	header := new(Header)

out:
	for {
		if exitError := header.Read(input); exitError != nil {
			break
		}
		logrus.Debug("received header: ", header)

		if header.Len > consensus.MaxMsgLen {
			exitError = errors.New("too big message size")
			break
		}

		// limit read
		rl := io.LimitReader(input, int64(header.Len))

		switch header.Type {
		case consensus.MsgTypePing:
			// update peer info & send Pong
			var msg Ping
			if exitError = msg.Read(rl); exitError != nil {
				break out
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
			p.queueMessage(resp)

		case consensus.MsgTypePong:
			// update peer info
			var msg Pong
			if exitError = msg.Read(rl); exitError != nil {
				break out
			}

			// update info
			p.Info.TotalDifficulty = msg.TotalDifficulty
			p.Info.Height = msg.Height

			logrus.Debug("received Pong: ", msg)

		case consensus.MsgTypeGetPeerAddrs:
			var msg GetPeerAddrs
			if exitError = msg.Read(rl); exitError != nil {
				break out
			}
			logrus.Info("received msgTypeGetPeerAddrs")

			// Send answer
			var resp PeerAddrs
			p.queueMessage(resp)

		case consensus.MsgTypePeerAddrs:
			var msg PeerAddrs
			if exitError = msg.Read(rl); exitError != nil {
				break out
			}
			logrus.Info("received msgTypePeerAddrs")
		case consensus.MsgTypeGetHeaders:
			logrus.Info("received msgTypeGetHeaders")
		case consensus.MsgTypeHeaders:
			logrus.Info("received msgTypeHeaders")
		case consensus.MsgTypeGetBlock:
			logrus.Info("received msgTypeGetBlock")
		case consensus.MsgTypeBlock:
			logrus.Info("received msgTypeBlock")
		case consensus.MsgTypeTransaction:
			logrus.Info("received msgTypeTransaction")

		default:
			exitError = errors.New("receive unexpected message (type) from peer")
			break out
		}

		// update recv bytes counter
		atomic.AddUint64(&p.bytesReceived, header.Len + consensus.HeaderLen)
	}

	p.wg.Done()
	p.Disconnect(exitError)
}

// Disconnect closes peer connection
func (p Peer) Disconnect(reason error) {
	logrus.Info("Disconnect peer: ", reason)

	close(p.quit)
	p.conn.Close()
	p.wg.Wait()
}

// SendPing sends Ping request to peer
func (p Peer) SendPing() {
	var request Ping
	request.TotalDifficulty = consensus.Difficulty(1)
	request.Height = 1

	p.queueMessage(request)
}
