package p2p

import (
	"bytes"
	"encoding/hex"
	"github.com/dblokhin/gringo/src/consensus"
	"testing"
	"time"
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

// TestHeaderParsing ensures a BlockHeader message is parsed correctly.
func TestHeaderParsing(t *testing.T) {
	// BlockHeader message for testnet4 block at height 65920.
	message, _ := hex.DecodeString("54340800000000000001950002000000000001e776086b8fa08b7f053690f9d0436b5279186d1a37ff21501187ff514a500c40d028000000005bb72a4f07c30daf06342f567ed56fdf1e014b8f06ec0d11d04129b7ca8023baabd8ad478e7b101d34cbc4e3de792173ff22c23b8e42147e17cbc51be737bd33382c59f281f249a756ea971a8750868b149b11a2d15eaa762f9e0787071c7d6282d025ed2649fb9d575beee4fb260fc9f2fb032e264b18b1e7e2a783ffd6b95d9f0e07d70815e90ad9bd8f6493a213a48680ea65ed988e1fee39e95d328509fa1695fc4e8100000000000438f2000000000003f6560000000003105eb00000000000000001b5a07fcfda228a4b1ed85c238189b12914ba6f859df5b9f0d863245ef1be049cd7d720594e1fabf6c145fdbbc6d4e8ca1a7378d42360307b127944d25843a9bad2685bc3c3ffeb99f386f13aa6b14b306096bd8adf0023fd5df0c7ec8bce18826fe5615049d346eaf88a49b84d6029524a358b4964a8dc34ba0d6dc6bd6ba5d8f8d5af1f435453eb9c80759d0b05bf259728b96f8e3ceff847ef4ccad69e1f21ec03")
	expectedHash, _ := hex.DecodeString("0f1a15c4a305815af54a30dd087352216b72e2893666750c400d31f73ff119d6")
	expectedPrevHash, _ := hex.DecodeString("086b8fa08b7f053690f9d0436b5279186d1a37ff21501187ff514a500c40d028")

	blockHeader := new(BlockHeader)
	read, err := ReadMessage(bytes.NewReader(message), blockHeader)

	if err != nil || read != uint64(len(message)) {
		t.Errorf("Failed to parse message: %v", err)
	}

	if read != uint64(len(message)) {
		t.Errorf("Failed to consume entire message: read %d of %d", read, len(message))
	}

	if bytes.Compare(expectedHash, blockHeader.Header.Hash()) != 0 {
		t.Error("Incorrect hash")
	}

	if blockHeader.Header.Version != 2 {
		t.Error("Incorrect version")
	}

	if blockHeader.Header.Height != 124790 {
		t.Error("Incorrect height")
	}

	if bytes.Compare(expectedPrevHash, blockHeader.Header.Previous) != 0 {
		t.Error("Incorrect previous block")
	}

	expectedTimestamp := time.Unix(1538730575, 0).UTC()
	if expectedTimestamp != blockHeader.Header.Timestamp {
		t.Errorf("Incorrect timestamp, got %v expected: %v", blockHeader.Header.Timestamp, expectedTimestamp)
	}

	UTXORoot := []byte{0x7, 0xc3, 0xd, 0xaf, 0x6, 0x34, 0x2f, 0x56, 0x7e, 0xd5, 0x6f, 0xdf, 0x1e, 0x1, 0x4b, 0x8f, 0x6, 0xec, 0xd, 0x11, 0xd0, 0x41, 0x29, 0xb7, 0xca, 0x80, 0x23, 0xba, 0xab, 0xd8, 0xad, 0x47}
	if bytes.Compare(UTXORoot, blockHeader.Header.UTXORoot) != 0 {
		t.Error("Incorrect UTXORoot")
	}

	RangeProofRoot := []byte{0x8e, 0x7b, 0x10, 0x1d, 0x34, 0xcb, 0xc4, 0xe3, 0xde, 0x79, 0x21, 0x73, 0xff, 0x22, 0xc2, 0x3b, 0x8e, 0x42, 0x14, 0x7e, 0x17, 0xcb, 0xc5, 0x1b, 0xe7, 0x37, 0xbd, 0x33, 0x38, 0x2c, 0x59, 0xf2}
	if bytes.Compare(RangeProofRoot, blockHeader.Header.RangeProofRoot) != 0 {
		t.Error("Incorrect RangeProofRoot")
	}

	KernelRoot := []byte{0x81, 0xf2, 0x49, 0xa7, 0x56, 0xea, 0x97, 0x1a, 0x87, 0x50, 0x86, 0x8b, 0x14, 0x9b, 0x11, 0xa2, 0xd1, 0x5e, 0xaa, 0x76, 0x2f, 0x9e, 0x7, 0x87, 0x7, 0x1c, 0x7d, 0x62, 0x82, 0xd0, 0x25, 0xed}
	if bytes.Compare(KernelRoot, blockHeader.Header.KernelRoot) != 0 {
		t.Error("Incorrect KernelRoot")
	}

	solution := []uint32{19094744, 21859404, 22802053, 24374075, 38157711, 39811247, 56581744, 65653540, 96597675, 104194026, 112376373, 128512230, 129172994, 153238665, 178589027, 191699543, 201311171, 204971215, 208244412, 213934231, 234400729, 251564416, 263397313, 272832977, 274851183, 305568330, 308461114, 314612592, 324314402, 338830533, 384357234, 397987233, 416851307, 419266223, 450155792, 451608889, 468734137, 479480722, 485636542, 501808925, 517392972, 526452988}
	for i, nonce := range blockHeader.Header.POW.Nonces {
		if solution[i] != nonce {
			t.Error("Incorrect proof of work")
		}
	}

	if blockHeader.Header.POW.EdgeBits != 30 {
		t.Error("Incorrect cuckoo size")
	}

	if blockHeader.Header.ScalingDifficulty != 1 {
		t.Error("Incorrect scaling difficulty")
	}

	if blockHeader.Header.TotalDifficulty != 51404464 {
		t.Error("Incorrect total difficulty")
	}

	totalKernelOffset, _ := hex.DecodeString("2649fb9d575beee4fb260fc9f2fb032e264b18b1e7e2a783ffd6b95d9f0e07d7")
	if bytes.Compare(totalKernelOffset, blockHeader.Header.TotalKernelOffset) != 0 {
		t.Error("Incorrect total kernel offset")
	}

	totalKernelSum := []byte{0x8, 0x15, 0xe9, 0xa, 0xd9, 0xbd, 0x8f, 0x64, 0x93, 0xa2, 0x13, 0xa4, 0x86, 0x80, 0xea, 0x65, 0xed, 0x98, 0x8e, 0x1f, 0xee, 0x39, 0xe9, 0x5d, 0x32, 0x85, 0x9, 0xfa, 0x16, 0x95, 0xfc, 0x4e, 0x81}
	if bytes.Compare(totalKernelSum, blockHeader.Header.TotalKernelSum) != 0 {
		t.Error("Incorrect total kernel sum")
	}

	if blockHeader.Header.Nonce != 13087601047833315915 {
		t.Error("Incorrect nonce")
	}

	if blockHeader.Header.OutputMmrSize != 0x438f2 {
		t.Error("Incorrect output mmr size")
	}

	if blockHeader.Header.KernelMmrSize != 0x3f656 {
		t.Error("Incorrect kernel mmr size")
	}

}
