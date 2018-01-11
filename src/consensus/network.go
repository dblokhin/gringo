package consensus

// MagicCode is expected in the header of every message
var MagicCode = [2]byte{0x1e, 0xc5}

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
	MsgTypeError        uint8 = iota
	MsgTypeHand
	MsgTypeShake
	MsgTypePing
	MsgTypePong
	MsgTypeGetPeerAddrs
	MsgTypePeerAddrs
	MsgTypeGetHeaders
	MsgTypeHeaders
	MsgTypeGetBlock
	MsgTypeBlock
	MsgTypeTransaction
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
	CapPeerList = 1 << 2
	CapFullNode = CapFullHist | CapUtxoHist | CapPeerList
)

// Network error codes
const (
	NetUnsupportedVersion int = 100
)