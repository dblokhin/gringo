// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package chain

import (
	"consensus"
	"sync"
	"container/list"
)

var Testnet1 = Chain{
	genesis: consensus.Block{},
}

type Chain struct {
	sync.RWMutex

	// Genesis block
	genesis consensus.Block

	blockHeaders list.List

}

// Genesis returns genesis block
func (c *Chain) Genesis() consensus.Block {
	return c.genesis
}

func (c *Chain) TotalDifficulty() consensus.Difficulty {
	return consensus.ZeroDifficulty
}

func (c *Chain) Height() uint64 {
	return 0
}

func (c *Chain) GetBlockHeaders(loc consensus.Locator) []consensus.BlockHeader {
	return nil
}

func (c *Chain) GetBlock(hash consensus.BlockHash) *consensus.Block {
	return nil
}

func (c *Chain) ProcessHeaders(headers []consensus.BlockHeader) error {
	return nil
}

func (c *Chain) ProcessBlock(block *consensus.Block) error {
	return nil
}