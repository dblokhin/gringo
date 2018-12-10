package secp256k1zkp

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"github.com/btcsuite/btcd/btcec"
	. "github.com/yoss22/bulletproofs"
	"math/big"
)

const (
	// TagPubkeyEven is prepended to a compressed pubkey to signal that the y
	// coordinate is even.
	TagPubkeyEven = 0x02

	// TagPubkeyOdd is prepended to a compressed pubkey to signal that the y
	// coordinate is odd.
	TagPubkeyOdd = 0x03
)

// RandomBytes returns 32 bytes of randomness.
func RandomBytes() [32]byte {
	buf := [32]byte{}
	_, err := rand.Read(buf[:])
	if err != nil {
		panic("Unable to generate random int")
	}

	return buf
}

// RandomInt returns a scalar from Z_n.
func RandomInt() *big.Int {
retry:
	buf := RandomBytes()

	r := &big.Int{}
	r.SetBytes(buf[:])

	if r.Cmp(btcec.S256().N) == 1 {
		goto retry
	}

	return r
}

// Signature is an argument of knowledge that the signer possesses a private
// key.
type Signature struct {
	S big.Int
	R Point
}

// Bytes serializes the signature.
func (s Signature) Bytes() [64]byte {
	var buf [64]byte
	Rx := GetB32(s.R.X)
	S := GetB32(&s.S)
	copy(buf[0:32], Rx[:])
	copy(buf[32:64], S[:])
	return buf
}

// SignMessage convinces a verifier in zero knowledge that they know
// the private key x for a public key P = x*G.
//
// The prover sends a random curve point R = k*G which acts as a blinding
// factor.
// The verifier issues a random challenge e.
// The prover returns s = k + ex (i.e. a blinded multiple of x).
// The verifier can now check
//   s*G == k*G + (ex)*G
//   s*G == R + e*P
func SignMessage(publicKey Point, privateKey big.Int, message [32]byte) Signature {
	// Compute a random nonce, k.
	k := RandomInt()

	// R is the public key for k.
	R := ScalarMulPoint(&G, k)

	// Compute a non-interactive challenge.
	Rx := GetB32(R.X)
	compressedPubkey := CompressPubkey(publicKey)
	challenge := ComputeHash(
		Rx[:],
		compressedPubkey[:],
		message[:])
	e := new(big.Int).SetBytes(challenge[:])

	s := Sum(k, Mul(e, &privateKey))

	return Signature{S: *s, R: *R}
}

// VerifySignature returns true if the given signature was computed by signing
// message with the private key for publicKey.
func VerifySignature(publicKey Point, message [32]byte, signature Signature) bool {
	Rx := GetB32(signature.R.X)
	compressedPubkey := CompressPubkey(publicKey)

	// Compute the non-interactive challenge.
	challenge := ComputeHash(
		Rx[:],
		compressedPubkey[:],
		message[:])
	e := new(big.Int).SetBytes(challenge[:])

	// Verify s*G == R + e*P.
	lhs := ScalarMulPoint(&G, &signature.S)
	rhs := SumPoints(&signature.R, ScalarMulPoint(&publicKey, e))

	return lhs.X.Cmp(rhs.X) == 0
}

// CommitValue returns the Pedersen commitment to the value v with blinding
// factor blind.
func CommitValue(blind, v *big.Int) *Point {
	return SumPoints(
		ScalarMulPoint(&G, blind),
		ScalarMulPoint(&H, v))
}

// CompressPubkey returns the point p as a 33-byte compressed pubkey.
func CompressPubkey(p Point) [33]byte {
	var buf [33]byte
	if p.Y.Bit(0) == 1 { // is odd
		buf[0] = TagPubkeyOdd
	} else {
		buf[0] = TagPubkeyEven
	}
	x := GetB32(p.X)
	copy(buf[1:33], x[:])
	return buf
}

// decompressPoint returns the y-coordinate for the given x coordinate.
func decompressPoint(xBytes []byte) *big.Int {
	x := new(big.Int).SetBytes(xBytes)

	// Derive the possible y coordinates from the secp256k1 curve
	// y² = x³ + 7.
	x3 := new(big.Int).Mul(x, x)
	x3.Mul(x3, x)
	x3.Add(x3, btcec.S256().Params().B)

	// y = ±sqrt(x³ + 7).
	y := ModSqrtFast(x3)

	return y
}

// DecodeSignature reads a 64-byte signature.
func DecodeSignature(signature [64]byte) Signature {
	s := new(big.Int).SetBytes(signature[32:64])

	R := new(Point)
	R.X = new(big.Int).SetBytes(signature[0:32])
	R.Y = decompressPoint(signature[0:32])

	return Signature{
		S: *s,
		R: *R,
	}
}

// ComputeHash returns the SHA256 hash of all of the inputs.
func ComputeHash(inputs ...[]byte) [32]byte {
	hasher := sha256.New()

	for i := range inputs {
		hasher.Write(inputs[i])
	}

	var result [32]byte
	copy(result[:], hasher.Sum(nil))
	return result
}

// ComputeMessage encodes fee and lockHeight into a 32-byte message.
func ComputeMessage(fee, lockHeight uint64) [32]byte {
	var msg [32]byte
	binary.BigEndian.PutUint64(msg[16:], fee)
	binary.BigEndian.PutUint64(msg[24:], lockHeight)
	return msg
}
