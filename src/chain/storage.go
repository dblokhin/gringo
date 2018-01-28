// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package chain

import "consensus"

// Storage represents storage methods for backends
// Storage doesnt check consensus rules!
// all errors in storage are fatals
type Storage interface {
	// Adding block to storage
	AddBlock(block *consensus.Block)
	// Del blocks from id and all of child
	DelBlock(id consensus.BlockID)
	// Returns full block by hash or height (or both)
	// if not found return nil
	GetBlock(id consensus.BlockID) *consensus.Block
	// returns head of blockchain
	GetLastBlock() *consensus.Block
	// Returns list of blocks from id
	From(id consensus.BlockID, limit int) consensus.BlockList
}

