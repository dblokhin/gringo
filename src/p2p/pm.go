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
)

// maxOnlineConnections should be override
// TODO: setting up by config
var (
	maxOnlineConnections = 15
	maxPeersTableSize    = 10000
)

// peerManager control connections with peers
type peerManager struct {
	sync.RWMutex
	connected int

	pool chan struct{}
	quit chan int

	PeersTable map[string]*peerInfo
}

// newPM returns PM instance
func newPM() *peerManager {
	pm := &peerManager{
		connected: 0,
		PeersTable: make(map[string]*peerInfo),
	}

	pm.pool = make(chan struct{}, maxOnlineConnections)
	pm.quit = make(chan int)

	return pm
}

// Ban closes connection & ban peer
func (pm *peerManager) Ban(addr string) {
	pm.Lock()
	peer, ok := pm.PeersTable[addr]
	if ok {
		peer.Status = psBanned
	}
	pm.Unlock()

	// Close connection
	peer.Peer.Close()
}

// IsBan returns true if addr is banned
func (pm *peerManager) IsBan(addr string) bool {
	pm.Lock()
	peer, ok := pm.PeersTable[addr]
	pm.Unlock()

	if ok {
		return peer.Status == psBanned
	}

	return false
}

// AddPeer adds new peer addr to pm
func (pm *peerManager) AddPeer(addr string) {
	// Don't add if big peer table
	if len(pm.PeersTable) > maxPeersTableSize {
		return
	}

	if netAddr, err := net.ResolveTCPAddr("tcp", addr); err != nil {
		// dont add invalid tcp addrs
		return
	} else {
		// FIXME: discard another IPs
		if netAddr.IP.IsMulticast() {
			return
		}
	}

	pm.Lock()
	defer pm.Unlock()

	// Checks for existing
	if _, ok := pm.PeersTable[addr]; ok {
		return
	}

	// Adds new
	pm.PeersTable[addr] = &peerInfo{
		psNew,
		nil,
		time.Unix(0, 0),
	}
}

// PeerAddrs returns peer list (no banned)
func (pm *peerManager) PeerAddrs(capabilities consensus.Capabilities) []*net.TCPAddr {
	result := make([]*net.TCPAddr, 0)

	// Getting peers randomly
	pm.Lock()
	for addr, v := range pm.PeersTable {
		// TODO: filter by capabilities

		if v.Status == psBanned {
			continue
		}

		if netAddr, err := net.ResolveTCPAddr("tcp", addr); err == nil {
			result = append(result, netAddr)
		} else {
			logrus.Error(err)
		}

		if len(result) == consensus.MaxPeerAddrs {
			break
		}
	}
	pm.Unlock()

	return result
}

// connectPeer connects peer from peerTable
func (pm *peerManager) connectPeer(addr string) error {
	// for empty string nonerror exit
	if len(addr) == 0 {
		return nil
	}

	pm.Lock()
	peer, ok := pm.PeersTable[addr]
	defer pm.Unlock()

	if !ok {
		return errors.New("peer doest exists in peerTable")
	}

	if peer.Status == psBanned || peer.Status == psConnected {
		logrus.Debug("dont connect to banned host (or already connected)")
		return nil
	}

	if pm.connected > maxOnlineConnections {
		return errors.New("too big online peers connections")
	}

	peerConn, err := NewPeer(addr)
	if err != nil {
		return err
	}

	pm.connected++

	pm.PeersTable[addr].Peer = peerConn
	pm.PeersTable[addr].Status = psConnected
	pm.PeersTable[addr].LastConn = time.Now()

	// And send ping / peers request
	peerConn.Start()
	peerConn.SendPing()
	peerConn.SendPeerRequest(consensus.CapFullNode)

	// on disconnect update info
	go func() {
		peerConn.WaitForDisconnect()

		pm.Lock()
		pm.connected--
		pm.PeersTable[addr].Status = psDisconnected
		pm.Unlock()

		<-pm.pool
	}()

	return nil
}

// Run starts network activity
func (pm *peerManager) Run() {

out:
	for {
		select {
		case <- pm.quit: break out

		case pm.pool <- struct{}{}:
			if err := pm.connectPeer(pm.notConnected()); err != nil {
				logrus.Error(err)
			}

			time.Sleep(time.Second)
		}
	}

	// Close all connections
	pm.Lock()
	for _, peer := range pm.PeersTable {
		peer.Peer.Close()
		peer.Status = psDisconnected
	}
	pm.Unlock()
}

// Close stops network activity
func (pm *peerManager) Close() {
	close(pm.quit)
}

// notConnected returns peer addr from table which not active
func (pm *peerManager) notConnected() string {
	pm.Lock()
	defer pm.Unlock()

	for addr, peer := range pm.PeersTable {
		if peer.Status == psNew || peer.Status == psDisconnected {
			return addr
		}
	}

	return ""
}

type peerInfo struct {
	Status peerStatus
	Peer   *Peer

	LastConn time.Time
}

type peerStatus int

const (
	psNew          peerStatus = iota
	psConnected
	psBanned
	psDisconnected
)
