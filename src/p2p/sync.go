// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package p2p

import (
)

var Syncher Sync

type Sync struct {
	// Peers manager
	PM *peerManager

	State SyncState
}

// Start starts sync proccess with initial peer addrs
func (s *Sync) Start(addrs []string) {
	s.PM = newPM()
	for _, addr := range addrs {
		s.PM.AddPeer(addr)
	}

	go s.PM.Run()
	<-s.PM.quit
}


type SyncState int

const (
	syncInit         SyncState = iota
	syncCollectPeers
)
