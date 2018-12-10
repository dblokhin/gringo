package secp256k1zkp

import (
	"bytes"
	"encoding/hex"
	. "github.com/yoss22/bulletproofs"
	"math/big"
	"testing"
)

func decompressPointFromHex(s string) *Point {
	point := new(Point)
	b, _ := hex.DecodeString(s)
	if err := point.Read(bytes.NewReader(b)); err != nil {
		panic(err)
	}
	return point
}

func DecodeHex64(s string) [64]byte {
	slice, _ := hex.DecodeString(s)
	var arr [64]byte
	copy(arr[:], slice)
	return arr
}

func TestVerifySignature(t *testing.T) {
	// Private key
	x := big.NewInt(8)

	// Public key for x.
	P := ScalarMulPoint(&G, x)

	msg := [32]byte{}

	// Create a signature for msg using the private key x.
	sig := SignMessage(*P, *x, msg)

	// Verify that msg was signed with the private key for P.
	if !VerifySignature(*P, msg, sig) {
		t.Errorf("failed to verify signature")
	}
}

func TestVerifyKernel(t *testing.T) {
	excess := decompressPointFromHex("092095ceab2c20f9a6109a7b0add8d488b3838dcc007c77a43cbe99a14a81b62e8")
	signature := DecodeHex64("804b2ed798221e8f4c139daeedeab487221be33db1adf9e129928564e1702b02fbbacaf4cbe4c4b122a9b39d2a7625b9254e43eeade171e9ccafda6dd8538acc")

	fee := uint64(2)
	lockHeight := uint64(0)

	msg := ComputeMessage(fee, lockHeight)

	sig := DecodeSignature(signature)

	if !VerifySignature(*excess, msg, sig) {
		t.Errorf("verify failed")
	}
}
