// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package consensus

// MagicCode is expected in the header of every message
var MagicCode = [2]byte{0x54, 0x34}

const (
	// protocolVersion version of grin p2p protocol
	ProtocolVersion uint32 = 1

	// size in bytes of a message header
	HeaderLen uint64 = 11

	// MaxMsgLen is the maximum size we're willing to accept for any message. Enforced by the
	// peer-to-peer networking layer only for DoS protection.
	MaxMsgLen uint64 = 20000000
)

// Types of p2p messages
const (
	MsgTypeError uint8 = iota
	MsgTypeHand
	MsgTypeShake
	MsgTypePing
	MsgTypePong
	MsgTypeGetPeerAddrs
	MsgTypePeerAddrs
	MsgTypeGetHeaders
	MsgTypeHeader
	MsgTypeHeaders
	MsgTypeGetBlock
	MsgTypeBlock
	MsgTypeGetCompactBlock
	MsgTypeCompactBlock
	MsgTypeStemTransaction
	MsgTypeTransaction
	MsgTypeTxHashSetRequest
	MsgTypeTxHashSetArchive
	MsgTypeBanReason
	MsgTypeGetTransaction
	MsgTypeTransactionKernel
)

// Capabilities of node
type Capabilities uint32

const (
	// We don't know (yet) what the peer can do.
	CapUnknown Capabilities = 0
	// Full archival node, has the whole history without any pruning.
	CapFullHist = 1 << 0
	// Can provide block headers and the UTXO set for some recent-enough height.
	CapUtxoHist = 1 << 1
	// Can provide a list of healthy peers
	CapPeerList     = 1 << 2
	CapFastSyncNode = CapUtxoHist | CapPeerList
	CapFullNode     = CapFullHist | CapUtxoHist | CapPeerList
)

// Network error codes
const (
	NetUnsupportedVersion int = 100
)

const (
	// Maximum number of hashes in a block header locator request
	MaxLocators int = 14

	// Maximum number of block headers a peer should ever send
	MaxBlockHeaders = 512

	// Maximum number of peer addresses a peer should ever send
	MaxPeerAddrs = 256
)

// Protocol defines grin-node network communicates
type Protocol interface {
	// TransmittedBytes bytes sent and received
	// TransmittedBytes() uint64

	// SendPing sends a Ping message to the remote peer. Will panic if handle has never
	// been called on this protocol.
	SendPing()

	// SendBlock sends a block to our remote peer
	SendBlock(block *Block)

	// Relays a transaction to the remote peer
	SendTransaction(tx Transaction)

	// Sends a request for block headers based on the provided block locator
	SendHeaderRequest(locator Locator)

	// Sends a request for a block from its hash
	SendBlockRequest(hash Hash)

	// Sends a request for some peer addresses
	SendPeerRequest(capabilities Capabilities)

	// Close the connection to the remote peer
	Close()
}
