// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

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
	"encoding/hex"
)

// Peer is a participant of p2p network
type Peer struct {
	conn net.Conn

	// The following fields are only meant to be used *atomically*
	bytesReceived uint64
	bytesSent     uint64

	quit      chan struct{}
	wg        sync.WaitGroup

	// Queue for sending message
	sendQueue chan Message

	// disconnect flag
	disconnect int32

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

	logrus.Infof("starting new peer (%s)", addr)
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	logrus.Infof("connected to peer (%s)", addr)
	shake, err := shakeByHand(conn)
	if err != nil {
		return nil, err
	}

	p := new(Peer)
	p.conn = conn
	p.quit = make(chan struct{})
	p.sendQueue = make(chan Message)

	p.Info.Version = shake.Version
	p.Info.Capabilities = shake.Capabilities
	p.Info.TotalDifficulty = shake.TotalDifficulty
	p.Info.UserAgent = shake.UserAgent

	return p, nil
}

// AcceptNewPeer creates peer accepting listening server conn
func AcceptNewPeer(conn net.Conn) (*Peer, error) {

	logrus.Info("accept new peer")
	hand, err := handByShake(conn)
	if err != nil {
		return nil, err
	}

	p := new(Peer)
	p.conn = conn
	p.quit = make(chan struct{})
	p.sendQueue = make(chan Message)

	p.Info.Version = hand.Version
	p.Info.Capabilities = hand.Capabilities
	p.Info.TotalDifficulty = hand.TotalDifficulty
	p.Info.UserAgent = hand.UserAgent

	return p, nil
}

// Start starts loop listening, write handler and so on
func (p *Peer) Start() {
	p.wg.Add(2)
	go p.writeHandler()
	go p.readHandler()
}

// writeHandler is a goroutine dedicated to reading messages off of an incoming
// queue, and writing them out to the wire.
//
// NOTE: This method MUST be run as a goroutine.
func (p *Peer) writeHandler() {
	var exitError error

out:
	for {
		select {
		case msg := <-p.sendQueue:
			// Ensure that conn is alive
			if atomic.LoadInt32(&p.disconnect) != 0 {
				break out
			}

			var written uint64
			if written, exitError = WriteMessage(p.conn, msg); exitError != nil {
				break out
			}

			atomic.AddUint64(&p.bytesSent, written)

		case <-p.quit:
			exitError = errors.New("peer exiting")
			break out
		}
	}

	p.wg.Done()
	p.Disconnect(exitError)
}

// WriteMessage places msg to send queue
func (p *Peer) WriteMessage(msg Message) {
	select {
	case <-p.quit: logrus.Info("cannot send message, peer is shutting down")
	case p.sendQueue <- msg:
	}
}

// readHandler is responsible for reading messages off the wire in series, then
// properly dispatching the handling of the message to the proper subsystem.
//
// NOTE: This method MUST be run as a goroutine.
func (p *Peer) readHandler() {
	var exitError error
	input := bufio.NewReader(p.conn)
	header := new(Header)

out:
	for atomic.LoadInt32(&p.disconnect) == 0 {
		if exitError = header.Read(input); exitError != nil {
			break out
		}

		if header.Len > consensus.MaxMsgLen {
			exitError = errors.New("too big message size")
			break out
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

			logrus.Debugf("received Ping (%s): %s", p.conn.RemoteAddr().String(), msg.String())
			// send Pong
			// TODO: send actual blockchain state
			var resp Pong
			resp.TotalDifficulty = consensus.Difficulty(1)
			resp.Height = 1
			p.WriteMessage(&resp)

		case consensus.MsgTypePong:
			// update peer info
			var msg Pong
			if exitError = msg.Read(rl); exitError != nil {
				break out
			}

			// update info
			p.Info.TotalDifficulty = msg.TotalDifficulty
			p.Info.Height = msg.Height

		case consensus.MsgTypeGetPeerAddrs:
			logrus.Infof("receiving peer request (%s)", p.conn.RemoteAddr().String())

			var msg GetPeerAddrs
			if exitError = msg.Read(rl); exitError != nil {
				break out
			}

			// Send answer
			var resp PeerAddrs
			resp.peers = Syncher.PM.PeerAddrs(msg.Capabilities)
			logrus.Debug("sending peers: ", resp.peers)

			p.WriteMessage(&resp)

		case consensus.MsgTypePeerAddrs:
			logrus.Infof("receiving peer addrs (%s)", p.conn.RemoteAddr().String())

			var msg PeerAddrs
			if exitError = msg.Read(rl); exitError != nil {
				break out
			}

			logrus.Infof("received %d peers", len(msg.peers))
			for _, p := range msg.peers {
				Syncher.PM.AddPeer(p.String())
			}


		case consensus.MsgTypeGetHeaders:
			logrus.Infof("receiving header request (%s)", p.conn.RemoteAddr().String())

			var msg GetBlockHeaders
			if exitError = msg.Read(rl); exitError != nil {
				break out
			}

			// response
			var resp BlockHeaders
			p.WriteMessage(&resp)

		case consensus.MsgTypeHeaders:
			logrus.Infof("receiving headers (%s)", p.conn.RemoteAddr().String())

			var msg BlockHeaders
			if exitError = msg.Read(rl); exitError != nil {
				break out
			}

			logrus.Debug("headers: ", msg.Headers)

		case consensus.MsgTypeGetBlock:
			logrus.Infof("receiving block request (%s)", p.conn.RemoteAddr().String())

			var msg GetBlockHash
			if exitError = msg.Read(rl); exitError != nil {
				break out
			}

			// TODO: Send answer & if not found do not send answer

		case consensus.MsgTypeBlock:
			logrus.Infof("receiving block (%s)", p.conn.RemoteAddr().String())

			var msg consensus.Block
			if exitError = msg.Read(rl); exitError != nil {
				break out
			}

			//logrus.Debug("block: ", msg)
			logrus.Info("block hash: ", hex.EncodeToString(msg.Header.Hash()))

		case consensus.MsgTypeTransaction:
			logrus.Infof("receiving transaction (%s)", p.conn.RemoteAddr().String())

			var msg consensus.Transaction
			if exitError = msg.Read(rl); exitError != nil {
				break out
			}

			logrus.Debug("transaction: ", msg)

		default:
			logrus.Debug("received unexpected message: ", header)
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
func (p *Peer) Disconnect(reason error) {
	if !atomic.CompareAndSwapInt32(&p.disconnect, 0, 1) {
		return
	}

	logrus.Info("Disconnect peer: ", reason)

	close(p.quit)
	p.conn.Close()
	p.wg.Wait()
}

// Close the connection to the remote peer
func (p *Peer) Close() {
	p.Disconnect(errors.New("closing peer"))
}

// WaitForDisconnect waits until the peer has disconnected.
func (p *Peer) WaitForDisconnect() {
	<-p.quit
	p.wg.Wait()
}

// SendPing sends Ping request to peer
func (p *Peer) SendPing() {
	logrus.Info("sending ping")

	var request Ping
	request.TotalDifficulty = consensus.Difficulty(1)
	request.Height = 1

	p.WriteMessage(&request)
}

// SendBlockRequest sends request block by hash
func (p *Peer) SendBlockRequest(hash consensus.BlockHash) {
	logrus.Info("sending block request")

	var request GetBlockHash
	request.Hash = hash

	logrus.Debug("block hash: ", hash)
	p.WriteMessage(&request)
}

// SendBlock sends Block to peer
func (p *Peer) SendBlock(block consensus.Block) {
	logrus.Info("sending block, height: ", block.Header.Height)
	p.WriteMessage(&block)
}

// SendPeerRequest sends peer request
func (p *Peer) SendPeerRequest(capabilities consensus.Capabilities) {
	logrus.Info("sending peer request")
	var request GetPeerAddrs

	request.Capabilities = capabilities

	p.WriteMessage(&request)
}

// SendHeaderRequest sends request headers
func (p *Peer) SendHeaderRequest(locator consensus.Locator) {
	logrus.Info("sending header request")

	if len(locator.Hashes) > consensus.MaxLocators {
		logrus.Debug("locator hashes count: ", len(locator.Hashes))
		logrus.Fatal(errors.New("too big locator hashes"))
	}

	var request GetBlockHeaders
	request.Locator = locator

	p.WriteMessage(&request)
}

// SendTransaction sends tx to peer
func (p *Peer) SendTransaction(tx consensus.Transaction) {
	logrus.Info("sending transaction")
	p.WriteMessage(&tx)
}
