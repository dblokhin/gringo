// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package p2p

import (
	"math/rand"
	"time"
)

const (
	noncesCap = 100
)

type nonceList struct{
	idx int
	list []uint64
}

// Init nonce list
func (n *nonceList) Init() {
	n.list = make([]uint64, noncesCap)

	for i := 0; i < noncesCap; i++ {
		n.list[i] = rand.Uint64()
	}
}

// NextNonce returns next nonce from list
func (n *nonceList) NextNonce() uint64{
	n.idx = (n.idx + 1) % noncesCap
	return n.list[n.idx]
}

func (n *nonceList) Consist(nonce uint64) bool {
	for _, v := range n.list {
		if nonce == v {
			return true
		}
	}

	return false
}

var serverNonces nonceList

func init() {
	// init rand
	rand.Seed(time.Now().UnixNano())

	// init nonce
	serverNonces.Init()
}