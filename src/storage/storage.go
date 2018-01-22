// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package storage

import "consensus"

// Storage represents storage methods for backends
// Storage doesnt check consensus rules!
type Storage interface {
	// Adding block to storage
	AddBlock(block *consensus.Block) error
	// Del blocks from id and all of child
	DelBlock(id BlockID) error
	// Returns full block by hash or height (or both)
	// if not found return nil
	GetBlock(id BlockID) (*consensus.Block, error)
	// Returns list of blocks from id
	From(id BlockID, limit int) (consensus.BlockList, error)
}

// BlockID identify block by Hash or/and Height (if not nill)
type BlockID struct {
	// Block hash
	Hash *consensus.BlockHash
	// Block height
	Height *uint64
}