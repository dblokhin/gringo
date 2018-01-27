// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

// mysql storage backend
// all errors in storage are fatals
package storage

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"consensus"
	"sync"
)

// NewSqlStorage returns blockchain Storage defined in /src/chain
func NewSqlStorage(db *sql.DB) *SqlStorage {
	return &SqlStorage{
		db: db,
	}
}

// SqlStorage sql storage backend for blockchain
type SqlStorage struct {
	sync.RWMutex

	// database instance
	db *sql.DB
}

// AddBlock adds block to storage
func (s *SqlStorage) AddBlock(block *consensus.Block) {

}

// DelBlock deletes blocks from id and all of child
func (s *SqlStorage) DelBlock(id consensus.BlockID) {

}

// GetBlock returns full block by hash or height (or both)
// if not found return nil
func (s *SqlStorage) GetBlock(id consensus.BlockID) *consensus.Block {
	return nil
}

// Returns list of blocks from id
func (s *SqlStorage) From(id consensus.BlockID, limit int) consensus.BlockList {
	return nil
}

// BlocksHashes returns hashes of blockchain
func (s *SqlStorage) BlocksHashes() []consensus.Hash {
	return nil
}

