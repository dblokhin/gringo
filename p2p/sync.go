// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package p2p

import (
	"github.com/dblokhin/gringo/consensus"
	"github.com/sirupsen/logrus"
)

type Blockchain interface {
	// Requite a mutex
	Lock()
	Unlock()
	RLock()
	RUnlock()

	Genesis() consensus.Block
	TotalDifficulty() consensus.Difficulty
	Height() uint64
	GetBlockHeaders(loc consensus.Locator) []consensus.BlockHeader
	GetBlock(hash consensus.Hash) *consensus.Block

	// ProcessHeaders processing block headers
	// Validate blockchain rules
	// ban peer with consensus error
	ProcessHeaders(headers []consensus.BlockHeader) error

	// ProcessBlock processing block
	// Validate blockchain rules
	// ban peer with consensus error
	// update Height & TotalDiff on peer from which recv the block

	// propagate block on new block to connected peer with less Height
	// clear tx's from pool on new block
	ProcessBlock(block *consensus.Block) error
}

type Mempool interface {
	// ProcessTx processing transaction
	// Validate blockchain rules
	// ban peer with consensus error
	ProcessTx(transaction *consensus.Transaction) error
}

type PeersPool interface {
	// PropagateBlock block to connected peer with less Height
	PropagateBlock(block *consensus.Block)

	// Peers returns live peers list (without banned)
	Peers(capabilities consensus.Capabilities) *PeerAddrs

	// PeerInfo returns peer structure
	PeerInfo(addr string) *peerInfo

	// Add peer
	Add(addr string)

	// Ban peer & ensure closed connection
	Ban(addr string)

	// Run & stop
	Run()
	Stop()
}

// Syncer synchronize blockchain & mempool via peers pool
type Syncer struct {
	// Chain is a grin blockchain
	Chain   Blockchain
	Mempool Mempool

	// Pool of peers (peers manager)
	Pool PeersPool
}

// Start starts sync proccess with initial peer addrs
func NewSyncer(addrs []string, chain Blockchain, mempool Mempool) *Syncer {

	sync := new(Syncer)
	sync.Chain = chain
	sync.Mempool = mempool
	sync.Pool = newPeersPool(sync)

	for _, addr := range addrs {
		sync.Pool.Add(addr)
	}

	return sync
}

// Run begins syncing with peers.
func (s *Syncer) Run() {
	s.Pool.Run()
}

// Stop stops activity
func (s *Syncer) Stop() {
	s.Pool.Stop()
}

func (s *Syncer) ProcessMessage(peer *Peer, message Message) {

	peerInfo := s.Pool.PeerInfo(peer.Addr)
	if peerInfo == nil {
		// should never rich
		logrus.Fatal("unexpected error")
	}

	switch msg := message.(type) {
	case *Ping:
		// MUST be answered
		// update peer info
		peerInfo.Lock()
		peerInfo.TotalDifficulty = msg.TotalDifficulty
		peerInfo.Height = msg.Height
		peerInfo.Unlock()

		// send answer
		var resp Pong

		// Lock the chain before getting various params
		s.Chain.RLock()
		resp.TotalDifficulty = s.Chain.TotalDifficulty()
		resp.Height = s.Chain.Height()
		s.Chain.RUnlock()

		peer.WriteMessage(&resp)
		logrus.Debugf("Sent Pong to %s", peer.conn.RemoteAddr())

	case *Pong:
		// update peer info
		peerInfo.Lock()
		peerInfo.TotalDifficulty = msg.TotalDifficulty
		peerInfo.Height = msg.Height
		peerInfo.Unlock()

		logrus.Debugf("Received Pong from %s", peer.conn.RemoteAddr())

	case *GetPeerAddrs:
		// MUST NOT be answered
		// Send answer
		peers := s.Pool.Peers(msg.Capabilities)
		if peers != nil {
			peer.WriteMessage(peers)
		}
		logrus.Debugf("Sent %d PeerAddrs to %s", len(peers.peers), peer.conn.RemoteAddr())

	case *PeerAddrs:
		// Adding peer to pool
		for _, p := range msg.peers {
			s.Pool.Add(p.String())
		}

	case *GetBlockHeaders:
		// MUST be answered
		// send answer
		headers := s.Chain.GetBlockHeaders(msg.Locator)
		resp := BlockHeaders{
			Headers: headers,
		}

		peer.WriteMessage(&resp)

	case *BlockHeader:
		headers := []consensus.BlockHeader{msg.Header}
		if err := s.Chain.ProcessHeaders(headers); err != nil {
			// ban peer ?
			//s.Pool.Ban(peer.conn.RemoteAddr().String())
			logrus.Infof("Failed to process header: %v", err)
		}

		logrus.Debugf("Received BlockHeader from %s for height %d: %v:", peer.conn.RemoteAddr(), msg.Header.Height, msg.Header.Hash())

	case *BlockHeaders:
		if err := s.Chain.ProcessHeaders(msg.Headers); err != nil {
			// ban peer ?
			s.Pool.Ban(peer.conn.RemoteAddr().String())
		}

	case *GetBlock:
		// MUST NOT be answered
		if block := s.Chain.GetBlock(msg.Hash); block != nil {
			peer.WriteMessage(block)
		}

	case *consensus.Block:
		// ProcessBlock puts block into blockchain
		// if block on the top of chain than propagate it
		// to others nodes with less TotalDifficulty
		if err := s.Chain.ProcessBlock(msg); err != nil {
			logrus.Info(err)
			// TODO: maybe smarter ban peer ?
			s.Pool.Ban(peer.conn.RemoteAddr().String())
		}

		// update peer info
		peerInfo.Lock()

		if peerInfo.TotalDifficulty < msg.Header.TotalDifficulty || peerInfo.Height < msg.Header.Height {
			peerInfo.TotalDifficulty = msg.Header.TotalDifficulty
			peerInfo.Height = msg.Header.Height
		}

		peerInfo.Unlock()

		// propagate if it is top block
		if msg.Header.Height == s.Chain.Height() {
			s.Pool.PropagateBlock(msg)
		}

	case *consensus.Transaction:
		if err := s.Mempool.ProcessTx(msg); err != nil {
			// ban peer ?
			s.Pool.Ban(peer.conn.RemoteAddr().String())
		}

		// TODO: propagate tx?
	}
}
