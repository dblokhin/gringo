// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package chain

import (
	"bytes"
	"errors"
	"github.com/dblokhin/gringo/src/consensus"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

// Testnet1 genesis block
var Testnet1 = consensus.Block{
	Header: consensus.BlockHeader{
		Version:         1,
		Height:          0,
		Previous:        bytes.Repeat([]byte{0xff}, consensus.BlockHashSize),
		Timestamp:       time.Date(2017, 11, 16, 20, 0, 0, 0, time.UTC),
		Difficulty:      10,
		TotalDifficulty: 10,

		UTXORoot:       bytes.Repeat([]byte{0x00}, 32),
		RangeProofRoot: bytes.Repeat([]byte{0x00}, 32),
		KernelRoot:     bytes.Repeat([]byte{0x00}, 32),

		Nonce: 28205,
		POW: consensus.Proof{
			Nonces: []uint32{
				0x21e, 0x7a2, 0xeae, 0x144e, 0x1b1c, 0x1fbd,
				0x203a, 0x214b, 0x293b, 0x2b74, 0x2bfa, 0x2c26,
				0x32bb, 0x346a, 0x34c7, 0x37c5, 0x4164, 0x42cc,
				0x4cc3, 0x55af, 0x5a70, 0x5b14, 0x5e1c, 0x5f76,
				0x6061, 0x60f9, 0x61d7, 0x6318, 0x63a1, 0x63fb,
				0x649b, 0x64e5, 0x65a1, 0x6b69, 0x70f8, 0x71c7,
				0x71cd, 0x7492, 0x7b11, 0x7db8, 0x7f29, 0x7ff8,
			},
		},
	},
}

// Testnet2 genesis block
var Testnet2 = consensus.Block{
	Header: consensus.BlockHeader{
		Version:   1,
		Height:    0,
		Previous:  bytes.Repeat([]byte{0xff}, consensus.BlockHashSize),
		Timestamp: time.Date(2017, 11, 16, 20, 0, 0, 0, time.UTC),
		//Difficulty:      10,
		//TotalDifficulty: 10,

		UTXORoot:       bytes.Repeat([]byte{0x00}, 32),
		RangeProofRoot: bytes.Repeat([]byte{0x00}, 32),
		KernelRoot:     bytes.Repeat([]byte{0x00}, 32),

		Nonce: 70081,
		POW: consensus.Proof{
			Nonces: []uint32{
				0x43ee48, 0x18d5a49, 0x2b76803, 0x3181a29, 0x39d6a8a, 0x39ef8d8,
				0x478a0fb, 0x69c1f9e, 0x6da4bca, 0x6f8782c, 0x9d842d7, 0xa051397,
				0xb56934c, 0xbf1f2c7, 0xc992c89, 0xce53a5a, 0xfa87225, 0x1070f99e,
				0x107b39af, 0x1160a11b, 0x11b379a8, 0x12420e02, 0x12991602, 0x12cc4a71,
				0x13d91075, 0x15c950d0, 0x1659b7be, 0x1682c2b4, 0x1796c62f, 0x191cf4c9,
				0x19d71ac0, 0x1b812e44, 0x1d150efe, 0x1d15bd77, 0x1d172841, 0x1d51e967,
				0x1ee1de39, 0x1f35c9b3, 0x1f557204, 0x1fbf884f, 0x1fcf80bf, 0x1fd59d40,
			},
		},
	},
}

// Mainnet genesis block
var Mainnet = consensus.Block{
	Header: consensus.BlockHeader{
		Version:         1,
		Height:          0,
		Previous:        bytes.Repeat([]byte{0xff}, consensus.BlockHashSize),
		Timestamp:       time.Date(2018, 8, 14, 0, 0, 0, 0, time.UTC),
		Difficulty:      1000,
		TotalDifficulty: 1000,

		UTXORoot:       bytes.Repeat([]byte{0x00}, 32),
		RangeProofRoot: bytes.Repeat([]byte{0x00}, 32),
		KernelRoot:     bytes.Repeat([]byte{0x00}, 32),

		Nonce: 28205,
		POW: consensus.Proof{
			Nonces: []uint32{
				0x21e, 0x7a2, 0xeae, 0x144e, 0x1b1c, 0x1fbd,
				0x203a, 0x214b, 0x293b, 0x2b74, 0x2bfa, 0x2c26,
				0x32bb, 0x346a, 0x34c7, 0x37c5, 0x4164, 0x42cc,
				0x4cc3, 0x55af, 0x5a70, 0x5b14, 0x5e1c, 0x5f76,
				0x6061, 0x60f9, 0x61d7, 0x6318, 0x63a1, 0x63fb,
				0x649b, 0x64e5, 0x65a1, 0x6b69, 0x70f8, 0x71c7,
				0x71cd, 0x7492, 0x7b11, 0x7db8, 0x7f29, 0x7ff8,
			},
		},
	},
}

type Chain struct {
	sync.RWMutex

	// Storage of blockchain
	storage Storage

	// genesis block
	genesis *consensus.Block
	// last block of blockchain
	head *consensus.Block
	// current height of chain
	height uint64
	// current total difficulty
	totalDifficulty consensus.Difficulty
}

func New(genesis *consensus.Block, storage Storage) *Chain {
	chain := Chain{
		storage:         storage,
		genesis:         genesis,
		head:            genesis,
		height:          genesis.Header.Height,
		totalDifficulty: genesis.Header.TotalDifficulty,
	}

	// init state from storage
	// setting up currents: height, total diff & blockHashChain
	if lastBlock := storage.GetLastBlock(); lastBlock != nil {
		chain.head = lastBlock
		chain.totalDifficulty = lastBlock.Header.TotalDifficulty
		chain.height = lastBlock.Header.Height
	}

	return &chain
}

// Genesis returns genesis block
func (c *Chain) Genesis() consensus.Block {
	return *c.genesis
}

// TotalDifficulty returns current total difficulty
func (c *Chain) TotalDifficulty() consensus.Difficulty {
	return c.totalDifficulty
}

// Height returns current height
func (c *Chain) Height() uint64 {
	return c.height
}

// GetBlockHeaders returns block headers
func (c *Chain) GetBlockHeaders(loc consensus.Locator) []consensus.BlockHeader {
	// for safety
	if len(loc.Hashes) > consensus.MaxLocators {
		logrus.Error("locator hashes object is too big")
		loc.Hashes = loc.Hashes[:consensus.MaxLocators]
	}

	result := make([]consensus.BlockHeader, 0)
	c.RLock()
	defer c.RUnlock()

	for _, hash := range loc.Hashes {

		// if hash is head of current chain, return empty result
		if bytes.Compare(hash, c.head.Hash()) == 0 {
			return result
		}

		blockID := consensus.BlockID{
			Hash:   hash,
			Height: nil,
		}

		// get blocks from
		blockList := c.storage.From(blockID, consensus.MaxBlockHeaders+1)
		if len(blockList) > 0 {
			// pass first block
			blockList = blockList[1:]

			// collect headers
			for _, block := range blockList {
				result = append(result, block.Header)
			}

			return result
		}
	}

	return result
}

// GetBlock returns block by hash, if not found returns nil, nil
func (c *Chain) GetBlock(hash consensus.Hash) *consensus.Block {
	if hash == nil {
		return nil
	}

	return c.storage.GetBlock(consensus.BlockID{
		Hash:   hash,
		Height: nil,
	})
}

// GetBlockID returns block by hash, height or both
func (c *Chain) GetBlockID(b consensus.BlockID) *consensus.Block {
	return c.storage.GetBlock(b)
}

func (c *Chain) ProcessHeaders(headers []consensus.BlockHeader) error {
	return nil
}

func (c *Chain) ProcessBlock(block *consensus.Block) error {
	// before locking storage on change MUST lock the Chain
	// Checking existing block
	c.Lock()
	defer c.Unlock()
	logrus.Infof("processing block (height: %d, totalDiff: %d)", block.Header.Height, block.Header.TotalDifficulty)

	// quick check is it current tip
	if bytes.Compare(c.head.Hash(), block.Hash()) == 0 {
		// the block is exists
		return nil
	}

	// verify block by consensus rules
	if err := block.Validate(); err != nil {
		return err
	}

	logrus.Info("getting the previous blocks")
	// Get the previous block
	prevHeight := block.Header.Height - 1
	prevBlockID := consensus.BlockID{
		Hash:   block.Header.Previous,
		Height: &prevHeight,
	}
	prevBlock := c.storage.GetBlock(prevBlockID)
	if prevBlock == nil {
		logrus.Info("no previous blocks")
		// No previous block at the current chain
		// It may be unknown fork-chain
		// TODO: process that
		return nil
	}

	logrus.Info("validating with the previous blocks")
	// Previous block exists

	// Checks with the previous block
	// - previous Timestamp MUST BE less block.Header.Timestamp
	if !block.Header.Timestamp.After(prevBlock.Header.Timestamp) {
		return errors.New("invalid block time")
	}
	// - block.TotalDiff MUST BE == previous.TotalDiff + previous.POW.ToDifficulty()
	if block.Header.TotalDifficulty != prevBlock.Header.TotalDifficulty+prevBlock.Header.POW.ToDifficulty() {
		return errors.New("wrong block total difficulty")
	}
	// - check that the difficulty is not less than that calculated by the
	//    	difficulty average based on the previous blocks
	limit := int(consensus.DifficultyAdjustWindow + consensus.MedianTimeWindow)
	fromHeight := uint64(0)
	if block.Header.Height > uint64(limit) {
		fromHeight = block.Header.Height - uint64(limit)
	}

	blockID := consensus.BlockID{
		Hash:   nil,
		Height: &fromHeight,
	}

	diffAvg := consensus.NextDifficulty(c.storage.From(blockID, limit))
	if block.Header.Difficulty < diffAvg {
		return errors.New("difficulty is too low")
	}

	// TODO: applying block

	return nil
}

// Head returns lastest block in blockchain
func (c *Chain) Head() consensus.Block {
	return *c.head
}

// Validate returns nil if chain successfully passed consensus rules
func (c *Chain) Validate() error {
	c.Lock()
	defer c.Unlock()

	block := c.head

	// go from head to genesis
	// TODO: MUST check all consensus rules
	for bytes.Compare(block.Header.Previous, c.genesis.Hash()) != 0 {
		err := block.Validate()
		if err == nil {
			return err
		}

		if block = c.GetBlock(block.Header.Previous); block == nil {
			return errors.New("invalid previous hash. blockchain integrity is broken")
		}

		// TODO: check other rules, may be checking block.Height?
	}

	return nil
}
