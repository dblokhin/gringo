package chain

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestSerialiseProof(t *testing.T) {
	pow := Testnet3.Header.POW.ProofBytes()
	expected, _ := hex.DecodeString("dc20a4c0fd1f19f80c3904144aefa0ced42ef6205bca8fc36b71e5c735539b318851b908f9844a2366726bd4e849938a67f4d4f2b0ba1af1c958192e508b00b36b11066bb2f1a6a7863ba2d7c00c3b9ed50dd1939ec185c58f41f1d51537b2db604c0c5c9559a5463afd13bf676d5917d5e17cb3efe06ff66d96f1a2bd6ee1b1f2df4d7b63bd0507f9b904b4472fe5c8f14c9b993ea48def03")

	if bytes.Compare(pow, expected) != 0 {
		t.Errorf("Genesis proof of work was %x wanted %x",
			pow, expected)
	}
}

func TestGenesisHash(t *testing.T) {
	hash := Testnet4.Hash()
	expected, _ := hex.DecodeString("0644cedb1acfdde4ee9e135ae61de3cbeb301b5f27a40a2c366da8e724292f20")

	if bytes.Compare(hash, expected) != 0 {
		t.Errorf("Genesis hash was %v wanted %x. Content:\n%x\n",
			hash, expected, Testnet4.Bytes())
	}
}
