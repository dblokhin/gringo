// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package p2p

import (
	"consensus"
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
	GetBlockHeaders(loc consensus.Locator) *BlockHeaders
	GetBlock(hash consensus.BlockHash) *consensus.Block

	// ProcessHeaders processing block headers
	// Validate blockchain rules
	// ban peer with consensus error
	ProcessHeaders(headers BlockHeaders) error

	// ProcessBlock processing block
	// Validate blockchain rules
	// ban peer with consensus error
	// update Height & TotalDiff on peer from which recv the block

	// propagate block on new block to connected peer with less Height
	// clear tx's from pool on new block
	ProcessBlock(block consensus.Block) error
}

type Mempool interface {
	// ProcessTx processing transaction
	// Validate blockchain rules
	// ban peer with consensus error
	ProcessTx(transaction consensus.Transaction) error
}

type PeersPool interface {
	// PropagateBlock block to connected peer with less Height
	PropagateBlock(block consensus.Block)

	// Peers returns live peers list (without banned)
	Peers(capabilities consensus.Capabilities) *PeerAddrs

	// PeerInfo returns peer structure
	PeerInfo(addr string) *peerInfo

	// Add peer
	Add(sync *Syncer, addr string)

	// Ban peer & ensure closed connection
	Ban(addr string)
}

type Syncer struct {
	Chain Blockchain
	Mempool Mempool
	Pool PeersPool

	// Pool of peers (peers manager)
	PM *peersPool
}

// Start starts sync proccess with initial peer addrs
func (s *Syncer) Start(addrs []string) {
	s.PM = NewPeersPool(s)
	for _, addr := range addrs {
		s.PM.Add(addr)
	}

	go s.PM.Run()
	<-s.PM.quit
}

func (s *Syncer) ProcessMessage(peer *Peer, message Message) {

	peerInfo := s.Pool.PeerInfo(peer.Addr)
	if peerInfo == nil {
		// should never rich
		logrus.Fatal("unexpected error")
	}

	switch msg := message.(type) {
	case Ping:
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

	case Pong:
		// update peer info
		peerInfo.Lock()
		peerInfo.TotalDifficulty = msg.TotalDifficulty
		peerInfo.Height = msg.Height
		peerInfo.Unlock()

	case GetPeerAddrs:
		// Send answer
		peers := s.Pool.Peers(msg.Capabilities)
		if peers != nil {
			peer.WriteMessage(peers)
		}

	case PeerAddrs:
		// Adding peer to pool
		for _, p := range msg.peers {
			s.Pool.Add(s, p.String())
		}

	case GetBlockHeaders:
		// send answer
		headers := s.Chain.GetBlockHeaders(msg.Locator)
		if headers != nil {
			peer.WriteMessage(headers)
		}

	case BlockHeaders:
		if err := s.Chain.ProcessHeaders(msg); err != nil {
			// ban peer ?
			s.Pool.Ban(peer.conn.RemoteAddr().String())
		}

	case GetBlock:
		block := s.Chain.GetBlock(msg.Hash)
		if block != nil {
			peer.WriteMessage(block)
		}

	case consensus.Block:
		// ProcessBlock puts block into blockchain
		// if block on the top of chain than propagate it
		// to others nodes with less TotalDifficulty
		if err := s.Chain.ProcessBlock(msg); err != nil {
			// ban peer ?
			s.Pool.Ban(peer.conn.RemoteAddr().String())
		}

		// propagate if it is top block
		if msg.Header.Height == s.Chain.Height() {
			s.Pool.PropagateBlock(msg)
		}

	case consensus.Transaction:
		if err := s.Mempool.ProcessTx(msg); err != nil {
			// ban peer ?
			s.Pool.Ban(peer.conn.RemoteAddr().String())
		}

		// TODO: propagate tx?
	}
}