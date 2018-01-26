// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package p2p

import (
	"sync"
	"time"
	"consensus"
	"errors"
	"github.com/sirupsen/logrus"
	"net"
	"sync/atomic"
	"fmt"
)

// maxOnlineConnections should be override
// TODO: setting up by config
var (
	maxOnlineConnections = 15
	maxPeersTableSize    = 10000
)

// NewPeersPool returns peers pool instance
func NewPeersPool(sync *Syncer) *peersPool {
	pp := &peersPool{
		connected:      0,
		sync:           sync,
		pool:           make(chan struct{}, maxOnlineConnections),
		quit:           make(chan int),
		PeersTable:     make(map[string]*peerInfo),
		ConnectedPeers: make(map[string]*peerInfo),
	}

	return pp
}

// peersPool control connections with peers
type peersPool struct {
	sync.Mutex

	connected int32
	sync      *Syncer

	pool chan struct{}
	quit chan int

	// all peers
	PeersTable map[string]*peerInfo

	// connected peers table
	ConnectedPeers map[string]*peerInfo
}

// Ban closes connection & ban peer
func (pp *peersPool) Ban(addr string) {
	pp.Lock()
	peer, ok := pp.PeersTable[addr]
	if ok {
		peer.Status = psBanned
	}
	pp.Unlock()

	// Close connection
	peer.Peer.Close()
}

// IsBan returns true if addr is banned
func (pp *peersPool) IsBan(addr string) bool {
	pp.Lock()
	peer, ok := pp.PeersTable[addr]
	pp.Unlock()

	if ok {
		return peer.Status == psBanned
	}

	return false
}

// AddPeer adds new peer addr to pm
func (pp *peersPool) Add(addr string) {
	// Don't add if big peer table
	if len(pp.PeersTable) > maxPeersTableSize {
		return
	}

	netAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		// dont add invalid tcp addrs
		return
	}

	// FIXME: discard another IPs
	if netAddr.IP.IsMulticast() {
		return
	}

	if netAddr.Port == 0 {
		return
	}

	pp.Lock()
	defer pp.Unlock()

	// Checks for existing
	if _, ok := pp.PeersTable[addr]; ok {
		return
	}

	// Adds new
	pp.PeersTable[addr] = &peerInfo{
		Status:          psNew,
		Peer:            nil,
		ProtocolVersion: 0,
		Height:          0,
		TotalDifficulty: consensus.ZeroDifficulty,
		Capabilities:    consensus.CapUnknown,
		LastConn:        time.Unix(0, 0),
	}
}

// PeerAddrs returns peer list (no banned)
func (pp *peersPool) Peers(capabilities consensus.Capabilities) *PeerAddrs {
	addrs := make([]*net.TCPAddr, 0)

	// Getting peers randomly
	pp.Lock()
	defer pp.Unlock()

	for addr, v := range pp.PeersTable {
		if v.Status == psBanned || v.Status == psFailedConn {
			continue
		}

		// filter by capabilities
		if (v.Capabilities & capabilities) != capabilities {
			continue
		}

		if netAddr, err := net.ResolveTCPAddr("tcp", addr); err == nil {
			addrs = append(addrs, netAddr)
		} else {
			logrus.Error(err)
		}

		if len(addrs) == consensus.MaxPeerAddrs {
			break
		}
	}

	return &PeerAddrs{
		peers: addrs,
	}
}

// PeerInfo returns peer structure
func (pp *peersPool) PeerInfo(addr string) *peerInfo {
	pp.Lock()
	peer, ok := pp.PeersTable[addr]
	pp.Unlock()
	if ok {
		return peer
	}

	return nil
}

// PropagateBlock propagates block to connected peers
func (pp *peersPool) PropagateBlock(block consensus.Block) {
	pp.Lock()
	defer pp.Unlock()

	for _, pi := range pp.ConnectedPeers {
		go func(peerInfo *peerInfo) {
			peerInfo.Lock()
			defer peerInfo.Unlock()

			// propagate if peer height or totalDiff less than newest block
			if peerInfo.Height < block.Header.Height || peerInfo.TotalDifficulty < block.Header.TotalDifficulty {
				if peer := peerInfo.Peer; peer != nil {
					peer.SendBlock(block)
				}
			}
		}(pi)
	}
}

// connectPeer connects peer from peerTable
func (pp *peersPool) connectPeer(addr string) error {
	// for empty string nonerror exit
	if len(addr) == 0 {
		return nil
	}

	pp.Lock()
	peer, ok := pp.PeersTable[addr]
	pp.Unlock()

	if !ok {
		return errors.New("peer doesn't exists at peersTable")
	}

	if peer.Status == psBanned || peer.Status == psConnected {
		logrus.Debug("dont connect to banned host (or already connected)")
		return nil
	}

	if atomic.LoadInt32(&pp.connected) > int32(maxOnlineConnections) {
		return errors.New("too big online peers connections")
	}

	peerConn, err := NewPeer(pp.sync, addr)
	if err != nil {
		pp.Lock()
		pp.PeersTable[addr].Status = psFailedConn
		pp.Unlock()
		return err
	}

	// Check the Protocol version
	if peerConn.Info.Version != consensus.ProtocolVersion {
		return fmt.Errorf("unexpected protocolVersion: %d", peerConn.Info.Version)
	}

	pp.Lock()
	pp.connected++

	// update peers table
	pp.PeersTable[addr].Peer = peerConn
	pp.PeersTable[addr].Status = psConnected
	pp.PeersTable[addr].LastConn = time.Now()

	pp.PeersTable[addr].ProtocolVersion = peerConn.Info.Version
	pp.PeersTable[addr].Height = peerConn.Info.Height
	pp.PeersTable[addr].TotalDifficulty = peerConn.Info.TotalDifficulty
	pp.PeersTable[addr].Capabilities = peerConn.Info.Capabilities

	// update connected peers
	pp.ConnectedPeers[addr] = pp.PeersTable[addr]
	pp.Unlock()

	// And send ping / peers request
	peerConn.Start()
	peerConn.SendPing()
	peerConn.SendPeerRequest(consensus.CapFullNode)

	// on disconnect update info
	go func() {
		peerConn.WaitForDisconnect()
		logrus.Info("closed connection")
		pp.Lock()
		pp.connected--
		// update peers & connected peers tables
		pp.PeersTable[addr].Status = psDisconnected
		delete(pp.ConnectedPeers, addr)
		pp.Unlock()

		<-pp.pool
	}()

	return nil
}

// Run starts network activity
func (pp *peersPool) Run() {

out:
	for {
		select {
		case <-pp.quit:
			break out

		case pp.pool <- struct{}{}:
			if err := pp.connectPeer(pp.notConnected()); err != nil {
				logrus.Error(err)
				<-pp.pool
			}

			time.Sleep(time.Second)
		}
	}

	// Close all connections
	pp.Lock()
	for _, peer := range pp.PeersTable {
		peer.Peer.Close()
		peer.Status = psDisconnected
	}
	pp.Unlock()
}

// Close stops network activity
func (pp *peersPool) Close() {
	close(pp.quit)
}

// notConnected returns peer addr from table which not active
func (pp *peersPool) notConnected() string {
	pp.Lock()
	defer pp.Unlock()

	// first, find good peers
	for addr, peer := range pp.PeersTable {
		if peer.Status == psNew || peer.Status == psDisconnected {
			return addr
		}
	}

	// second, try to open conn with failed nodes
	for addr, peer := range pp.PeersTable {
		if peer.Status == psFailedConn {
			return addr
		}
	}

	return ""
}

type peerInfo struct {
	sync.Mutex

	Status peerStatus
	Peer   *Peer

	ProtocolVersion uint32
	Height          uint64
	TotalDifficulty consensus.Difficulty
	Capabilities    consensus.Capabilities

	LastConn time.Time
}

type peerStatus int

const (
	psNew          peerStatus = iota
	psConnected
	psBanned
	psDisconnected
	psFailedConn
)
