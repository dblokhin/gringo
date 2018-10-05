package p2p

import (
	"bytes"
	"encoding/hex"
	"github.com/dblokhin/gringo/src/chain"
	"github.com/dblokhin/gringo/src/consensus"
	"testing"
)

// TestShakeParsing ensures a shake message is parsed correctly.
func TestShakeParsing(t *testing.T) {
	message, _ := hex.DecodeString("1ec502000000000000004500000001000000060000000000007530000000000000000d4d572f4772696e20302e332e30006642e037a073b89c00f48c93cf1b701b335dab84b538136f63b6feadaa50d6")

	sh := new(shake)
	read, err := ReadMessage(bytes.NewReader(message), sh)
	if err != nil {
		t.Errorf("ReadMessage failed")
	}

	if read != uint64(len(message)) {
		t.Errorf("Failed to consume entire message: read %d of %d", read, len(message))
	}

	if sh.Version != 1 {
		t.Errorf("Version differs")
	}

	if sh.Capabilities != consensus.CapFastSyncNode {
		t.Errorf("Capabilities differs got %d", sh.Capabilities)
	}

	if sh.TotalDifficulty != 30000 {
		t.Errorf("TotalDifficulty differs")
	}

	if sh.UserAgent != "MW/Grin 0.3.0" {
		t.Errorf("Version differs")
	}

	if bytes.Compare(sh.Genesis, chain.Testnet3.Hash()) != 0 {
		t.Errorf("Genesis differs")
	}
}

// TestHeaderParsing ensures a BlockHeader message is parsed correctly.
func TestHeaderParsing(t *testing.T) {
	message, _ := hex.DecodeString("1ec50800000000000001950002000000000001e7000b606c40959ab5b7603dbc598a8b38efe4aa0ca02efe52cd649d6f99c75116aa000000005bb70c76eda4bb935244646249996b81f764bccbb9a298a506cdfef65c56dc4c8755f9177b87dfd2927abab336f305213037bc2665317b1723123173cd48e1947c1f037c1db7e25bdfc69ff6f416bab8b45be5b5a04b4d48fa8f1f55c3df27cf6a3c473e2649fb9d575beee4fb260fc9f2fb032e264b18b1e7e2a783ffd6b95d9f0e07d70996a21513286fecaa235250b5c746f11ffe486640ba2a085c9e2724981dab5d32000000000004380b000000000003f56800000000030fd31c0000000000000001d50f7b3617a421dd1ee61e2bc0d2c90c2c49e6877c5519f18bba282cf11e05246db1c0dc182005c41a2479068f9c72c09e2c91d2e4189eb5f65a3257d7fef90a08c560b26cfe6cab89a385f6c5bcce19c1a76dfcf98ec92fe0bde30e1c33f787810fd931f0ac4aba2dd9d1a21bc58ae27d5b4b902f2c307e8cb52c55b27959f816ee9cf02e0cbb5eea970c7cefaa9df158cf33946308673ee2ec6d4140be4388f703")

	blockHeader := new(BlockHeader)
	read, err := ReadMessage(bytes.NewReader(message), blockHeader)

	if err != nil || read != uint64(len(message)) {
		t.Errorf("Failed to parse message: %v", err)
	}

	if read != uint64(len(message)) {
		t.Errorf("Failed to consume entire message: read %d of %d", read, len(message))
	}
}
