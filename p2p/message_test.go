package p2p

import (
	"bytes"
	"encoding/hex"
	"github.com/dblokhin/gringo/src/consensus"
	"testing"
)

// TestShakeParsing ensures a shake message is parsed correctly.
func TestShakeParsing(t *testing.T) {
	message, _ := hex.DecodeString("543402000000000000004500000001000000060000000000007530000000000000000d4d572f4772696e20302e332e30006642e037a073b89c00f48c93cf1b701b335dab84b538136f63b6feadaa50d6")

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
}
