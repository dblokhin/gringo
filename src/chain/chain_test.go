package chain

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestGenesisHash(t *testing.T) {
	hash := Testnet4.Hash()
	expected, _ := hex.DecodeString("0644cedb1acfdde4ee9e135ae61de3cbeb301b5f27a40a2c366da8e724292f20")

	if bytes.Compare(hash, expected) != 0 {
		t.Errorf("Genesis hash was %v wanted %x. Content:\n%x\n",
			hash, expected, Testnet4.Bytes())
	}
}
