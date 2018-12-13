// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package p2p

import (
	"errors"
	"fmt"
	"github.com/dblokhin/gringo/src/consensus"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

// maxOnlineConnections should be override
// TODO: setting up by config
var (
	maxOnlineConnections = 15
	maxPeersTableSize    = 10000
)

// newPeersPool returns peers pool instance
func newPeersPool(sync *Syncer) *peersPool {
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
	ptmu sync.Mutex // mutex for PeersTable
	cpmu sync.Mutex // mutex for ConnectedPeers
	bnmu sync.Mutex // mutex for BannedPeers

	connected int32
	sync      *Syncer

	pool chan struct{}
	quit chan int

	// all peers
	PeersTable map[string]*peerInfo

	// connected peers table
	ConnectedPeers map[string]*peerInfo

	// banned peers
	BannedPeers map[string]struct{}
}

// Ban closes connection & ban peer
func (pp *peersPool) Ban(addr string) {
	pp.ptmu.Lock()
	peerInfo, ok := pp.PeersTable[addr]
	pp.ptmu.Unlock()

	if !ok {
		return
	}

	// Mark banned & Close connection
	peerInfo.Status = psBanned
	peerInfo.Peer.Close()

	// Add to ban list
	pp.bnmu.Lock()
	pp.BannedPeers[addr] = struct{}{}
	pp.bnmu.Unlock()

	// Clear the peers table
	pp.ptmu.Lock()
	delete(pp.PeersTable, addr)
	pp.ptmu.Unlock()
}

// IsBan returns true if addr is banned
func (pp *peersPool) IsBan(addr string) bool {
	pp.bnmu.Lock()
	defer pp.bnmu.Unlock()

	_, ok := pp.BannedPeers[addr]
	return ok
}

// AddPeer adds new peer addr to pm
func (pp *peersPool) Add(addr string) {
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

	pp.ptmu.Lock()
	defer pp.ptmu.Unlock()

	// Don't add if big peer table
	if len(pp.PeersTable) > maxPeersTableSize {
		return
	}

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
	pp.ptmu.Lock()
	defer pp.ptmu.Unlock()

	for addr, peerInfo := range pp.PeersTable {
		if peerInfo.Status == psBanned || peerInfo.Status == psFailedConn {
			continue
		}

		// filter by capabilities
		if (peerInfo.Capabilities & capabilities) != capabilities {
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
	pp.ptmu.Lock()
	peerInfo, ok := pp.PeersTable[addr]
	pp.ptmu.Unlock()
	if ok {
		return peerInfo
	}

	return nil
}

// PropagateBlock propagates block to connected peers
func (pp *peersPool) PropagateBlock(block *consensus.Block) {
	pp.cpmu.Lock()
	defer pp.cpmu.Unlock()

	for _, pi := range pp.ConnectedPeers {
		go func(peerInfo *peerInfo) {
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

	if pp.connected > int32(maxOnlineConnections) {
		return errors.New("too big online peers connections")
	}

	pp.ptmu.Lock()
	peerInfo, ok := pp.PeersTable[addr]
	pp.ptmu.Unlock()

	if !ok {
		return errors.New("peer doesn't exists at peersTable")
	}

	peerInfo.Lock()
	defer peerInfo.Unlock()

	if peerInfo.Status == psBanned || peerInfo.Status == psConnected {
		logrus.Debug("dont connect to banned host (or already connected)")
		return nil
	}

	peerConn, err := NewPeer(pp.sync, addr)
	if err != nil {
		peerInfo.Status = psFailedConn
		return err
	}

	// Check the Protocol version
	if peerConn.Info.Version != consensus.ProtocolVersion {
		return fmt.Errorf("unexpected protocolVersion: %d", peerConn.Info.Version)
	}

	pp.connected++

	// update peers table
	peerInfo.Peer = peerConn
	peerInfo.Status = psConnected
	peerInfo.LastConn = time.Now()

	peerInfo.ProtocolVersion = peerConn.Info.Version
	peerInfo.Height = peerConn.Info.Height
	peerInfo.TotalDifficulty = peerConn.Info.TotalDifficulty
	peerInfo.Capabilities = peerConn.Info.Capabilities

	// update connected peers
	pp.cpmu.Lock()
	pp.ConnectedPeers[addr] = peerInfo
	pp.cpmu.Unlock()

	// And send ping / peers request
	peerConn.Start()
	peerConn.SendPing()
	peerConn.SendPeerRequest(consensus.CapFullNode)

	// on disconnect update info
	go func() {
		peerConn.WaitForDisconnect()
		logrus.Infof("closed peer connection (%s)", addr)

		// update peers & connected peers tables
		peerInfo.Lock()
		peerInfo.Status = psDisconnected
		peerInfo.Unlock()

		// clean connected peers
		pp.connected--
		pp.cpmu.Lock()
		delete(pp.ConnectedPeers, addr)
		pp.cpmu.Unlock()

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
	pp.ptmu.Lock()
	defer pp.ptmu.Unlock()

	for _, pi := range pp.PeersTable {
		go func(peerInfo *peerInfo) {
			peerInfo.Lock()
			peerInfo.Peer.Close()
			peerInfo.Status = psDisconnected
			peerInfo.Unlock()
		}(pi)

	}
}

// Stop stops network activity
func (pp *peersPool) Stop() {
	close(pp.quit)
}

// notConnected returns peer addr from table which not active
func (pp *peersPool) notConnected() string {
	pp.ptmu.Lock()
	defer pp.ptmu.Unlock()

	// first, find good peers
	for addr, peerInfo := range pp.PeersTable {
		if peerInfo.Status == psNew || peerInfo.Status == psDisconnected {
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
	psNew peerStatus = iota
	psConnected
	psBanned
	psDisconnected
	psFailedConn
)
