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

	PeersTable map[string]*peerInfo
}

// newPM returns PM instance
func newPM() *peerManager {
	return &peerManager{
		connected: 0,
		PeersTable: make(map[string]*peerInfo),
	}
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
func (pm *peerManager) PeerAddrs() []string {
	result := make([]string, consensus.MaxPeerAddrs)
	cnt := 0;

	// Getting peers randomly
	pm.Lock()
	for k, v := range pm.PeersTable {
		if cnt == consensus.MaxPeerAddrs {
			break
		}

		if v.Status != psBanned {
			cnt++
			result = append(result, k)
		}
	}
	pm.Unlock()

	return result
}

// connectPeer connects peer from peerTable
func (pm *peerManager) connectPeer(addr string) error {
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

	// And send peers request
	peerConn.SendPeerRequest(consensus.CapFullNode)

	// on disconnect update info
	go func() {
		peerConn.WaitForDisconnect()

		pm.Lock()
		pm.connected--
		pm.PeersTable[addr].Status = psDisconnected
		pm.Unlock()
	}()

	return nil
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
