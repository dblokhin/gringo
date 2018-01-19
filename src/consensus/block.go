// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package consensus

import (
	"io"
	"time"
	"bytes"
	"encoding/binary"
	"github.com/sirupsen/logrus"
	"secp256k1zkp"
	"errors"
	"golang.org/x/crypto/blake2b"
	"fmt"
)

// BlockHash is hash of block (32 byte)
type BlockHash []byte
type Hash []byte

// SwitchCommitHashSize The size to use for the stored blake2 hash of a switch_commitment
const SwitchCommitHashSize = 20

// OutputFeatures is options for block validation
type OutputFeatures uint8

const (
	// No flags
	DefaultOutput OutputFeatures = 0
	// Output is a coinbase output, must not be spent until maturity
	CoinbaseOutput OutputFeatures = 1 << 0
)

// KernelFeatures is options for a kernel's structure or use
type KernelFeatures uint8

const (
	// No flags
	DefaultKernel KernelFeatures = 0
	// Kernel matching a coinbase output
	CoinbaseKernel KernelFeatures = 1 << 0
)

// Block of grin blockchain
type Block struct {
	Header BlockHeader

	Inputs  []Input
	Outputs []Output
	Kernels []TxKernel
}

// Bytes implements p2p Message interface
func (b *Block) Bytes() []byte {
	// Dump header
	buff := new(bytes.Buffer)
	if _, err := buff.Write(b.Header.Bytes()); err != nil {
		logrus.Fatal(err)
	}

	// Write counts: inputs, outputs, kernels
	if err := binary.Write(buff, binary.BigEndian, uint64(len(b.Inputs))); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, uint64(len(b.Outputs))); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, uint64(len(b.Kernels))); err != nil {
		logrus.Fatal(err)
	}

	// Write inputs
	for _, input := range b.Inputs {
		if _, err := buff.Write(input.Commit); err != nil {
			logrus.Fatal(err)
		}
	}

	// Write outputs
	for _, output := range b.Outputs {
		if _, err := buff.Write(output.Bytes()); err != nil {
			logrus.Fatal(err)
		}
	}

	// Write kernels
	for _, txKernel := range b.Kernels {
		if _, err := buff.Write(txKernel.Bytes()); err != nil {
			logrus.Fatal(err)
		}
	}

	return buff.Bytes()
}

// Type implements p2p Message interface
func (b *Block) Type() uint8 {
	return MsgTypeBlock
}

// Read implements p2p Message interface
func (b *Block) Read(r io.Reader) error {
	// Read block header
	logrus.Info("Read block header")
	if err := b.Header.Read(r); err != nil {
		return err
	}

	// Read counts
	var inputs, outputs, kernels uint64
	if err := binary.Read(r, binary.BigEndian, &inputs); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &outputs); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &kernels); err != nil {
		return err
	}

	logrus.Debugf("block inputs/outputs/kernels: %d, %d, %d", inputs, outputs, kernels)

	// Read inputs
	logrus.Info("Read inputs")
	b.Inputs = make([]Input, inputs)
	for i := uint64(0); i < inputs; i++ {
		commitment := make([]byte, secp256k1zkp.PedersenCommitmentSize)
		if _, err := io.ReadFull(r, commitment); err != nil {
			return err
		}

		b.Inputs[i].Commit = commitment
	}

	logrus.Debug("block inputs: ", b.Inputs)

	// Read outputs
	logrus.Info("Read outputs")
	b.Outputs = make([]Output, outputs)
	for i := uint64(0); i < outputs; i++ {
		if err := b.Outputs[i].Read(r); err != nil {
			return err
		}
	}

	logrus.Debug("block outputs: ", b.Outputs)

	// Read kernels
	logrus.Info("Read kernels")
	b.Kernels = make([]TxKernel, kernels)
	for i := uint64(0); i < kernels; i++ {
		if err := b.Kernels[i].Read(r); err != nil {
			return err
		}
	}

	logrus.Debug("block kernels: ", b.Kernels)

	return nil
}

// String implements String() interface
func (p Block) String() string {
	return fmt.Sprintf("%#v", p)
}

type Input struct {
	Commit secp256k1zkp.Commitment
}

// Output for a transaction, defining the new ownership of coins that are being
// transferred. The commitment is a blinded value for the output while the
// range proof guarantees the commitment includes a positive value without
// overflow and the ownership of the private key. The switch commitment hash
// provides future-proofing against quantum-based attacks, as well as provides
// wallet implementations with a way to identify their outputs for wallet
// reconstruction
//
// The hash of an output only covers its features, lock_height, commitment,
// and switch commitment. The range proof is expected to have its own hash
// and is stored and committed to separately.
type Output struct {
	// Options for an output's structure or use
	Features OutputFeatures
	// The homomorphic commitment representing the output's amount
	Commit secp256k1zkp.Commitment
	// The switch commitment hash, a 160 bit length blake2 hash of blind*J
	SwitchCommitHash SwitchCommitHash
	// A proof that the commitment is in the right range
	RangeProof secp256k1zkp.RangeProof
}

// Bytes implements p2p Message interface
func (o *Output) Bytes() []byte {
	buff := new(bytes.Buffer)

	// Write features
	if err := binary.Write(buff, binary.BigEndian, uint8(o.Features)); err != nil {
		logrus.Fatal(err)
	}

	// Write commitment
	if len(o.Commit) != secp256k1zkp.PedersenCommitmentSize {
		logrus.Fatal(errors.New("invalid input commitment len"))
	}

	if _, err := buff.Write(o.Commit); err != nil {
		logrus.Fatal(err)
	}

	// Write SwitchCommitHash
	if len(o.SwitchCommitHash) != SwitchCommitHashSize {
		logrus.Fatal(errors.New("invalid input switchCommitHash len"))
	}

	if _, err := buff.Write(o.SwitchCommitHash); err != nil {
		logrus.Fatal(err)
	}

	// Write range proof
	if len(o.RangeProof.Proof) > int(secp256k1zkp.MaxProofSize) || len(o.RangeProof.Proof) != o.RangeProof.ProofLen {
		logrus.Fatal(errors.New("invalid range proof len"))
	}

	if err := binary.Write(buff, binary.BigEndian, uint64(o.RangeProof.ProofLen)); err != nil {
		logrus.Fatal(err)
	}

	if _, err := buff.Write(o.RangeProof.Proof); err != nil {
		logrus.Fatal(err)
	}

	return buff.Bytes()
}

// Read implements p2p Message interface
func (o *Output) Read(r io.Reader) error {
	// Read features
	if err := binary.Read(r, binary.BigEndian, (*uint8)(&o.Features)); err != nil {
		return err
	}

	// Read commitment
	commitment := make([]byte, secp256k1zkp.PedersenCommitmentSize)
	if _, err := io.ReadFull(r, commitment); err != nil {
		return err
	}

	o.Commit = commitment

	// Read SwitchCommitHash
	hash := make([]byte, SwitchCommitHashSize)
	if _, err := io.ReadFull(r, hash); err != nil {
		return err
	}

	o.SwitchCommitHash = hash

	// Read range proof
	var proofLen uint64 // tha max is MaxProofSize (5134), but in message field it is uint64
	if err := binary.Read(r, binary.BigEndian, &proofLen); err != nil {
		return err
	}

	if proofLen > uint64(secp256k1zkp.MaxProofSize) {
		return errors.New("invalid range proof len")
	}

	proof := make([]byte, proofLen)
	if _, err := io.ReadFull(r, proof); err != nil {
		return err
	}

	o.RangeProof = secp256k1zkp.RangeProof{
		Proof:    proof,
		ProofLen: int(proofLen),
	}

	return nil
}

// String implements String() interface
func (p Output) String() string {
	return fmt.Sprintf("%#v", p)
}

// SwitchCommitHash the switch commitment hash
type SwitchCommitHash []byte // size = const SwitchCommitHashSize

// A proof that a transaction sums to zero. Includes both the transaction's
// Pedersen commitment and the signature, that guarantees that the commitments
// amount to zero.
// The signature signs the Fee and the LockHeight, which are retained for
// signature validation.
type TxKernel struct {
	// Options for a kernel's structure or use
	Features KernelFeatures
	// Fee originally included in the transaction this proof is for.
	Fee uint64
	// This kernel is not valid earlier than lockHeight blocks
	// The max lockHeight of all *inputs* to this transaction
	LockHeight uint64
	// Remainder of the sum of all transaction commitments. If the transaction
	// is well formed, amounts components should sum to zero and the excess
	// is hence a valid public key.
	Excess secp256k1zkp.Commitment
	// The signature proving the excess is a valid public key, which signs
	// the transaction fee.
	ExcessSig []byte
}

// Read implements p2p Message interface
func (k *TxKernel) Bytes() []byte {
	buff := new(bytes.Buffer)

	// Write features, fee & lock
	if err := binary.Write(buff, binary.BigEndian, uint8(k.Features)); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, k.Fee); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, k.LockHeight); err != nil {
		logrus.Fatal(err)
	}

	// Write Excess
	if len(k.Excess) != secp256k1zkp.PedersenCommitmentSize {
		logrus.Fatal(errors.New("invalid excess len"))
	}

	if _, err := buff.Write(k.Excess); err != nil {
		logrus.Fatal(err)
	}

	// Write ExcessSig
	if len(k.ExcessSig) > secp256k1zkp.MaxSignatureSize {
		logrus.Fatal(errors.New("invalid excess_sig len"))
	}
	if err := binary.Write(buff, binary.BigEndian, uint64(len(k.ExcessSig))); err != nil {
		logrus.Fatal(err)
	}

	if _, err := buff.Write(k.ExcessSig); err != nil {
		logrus.Fatal(err)
	}

	return buff.Bytes()
}

// Read implements p2p Message interface
func (k *TxKernel) Read(r io.Reader) error {
	// Read features, fee & lock
	if err := binary.Read(r, binary.BigEndian, (*uint8)(&k.Features)); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &k.Fee); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &k.LockHeight); err != nil {
		return err
	}

	// Read Excess
	commitment := make([]byte, secp256k1zkp.PedersenCommitmentSize)
	if _, err := io.ReadFull(r, commitment); err != nil {
		return err
	}

	k.Excess = commitment

	var excessSigLen uint64
	if err := binary.Read(r, binary.BigEndian, &excessSigLen); err != nil {
		return err
	}

	if excessSigLen > uint64(secp256k1zkp.MaxSignatureSize) {
		return errors.New("invalid excess_sig len")
	}

	k.ExcessSig = make([]byte, excessSigLen)
	if _, err := io.ReadFull(r, k.ExcessSig); err != nil {
		return err
	}

	return nil
}

// String implements String() interface
func (p TxKernel) String() string {
	return fmt.Sprintf("%#v", p)
}

// BlockHeader header of the grin-blocks
type BlockHeader struct {
	// Version of the block
	Version uint16
	// Height of this block since the genesis block (height 0)
	Height uint64
	// Hash of the block previous to this in the chain
	Previous BlockHash
	// Timestamp at which the block was built
	Timestamp time.Time
	// UTXORoot Merklish root of all the commitments in the UTXO set
	UTXORoot Hash
	// RangeProofRoot Merklish root of all range proofs in the UTXO set
	RangeProofRoot Hash
	// Merklish root of all transaction kernels in the UTXO set
	KernelRoot Hash
	// Nonce increment used to mine this block
	Nonce uint64
	// RangeProof of work data.
	POW Proof
	// Difficulty used to mine the block.
	Difficulty Difficulty
	// Total accumulated difficulty since genesis block
	TotalDifficulty Difficulty
}

// Bytes implements p2p Message interface
func (b *BlockHeader) Hash() []byte {
	hash := blake2b.Sum256(b.bytesWithoutPOW())

	return hash[:]
}

// bytesWithoutPOW used in Hash() method, where doesnt need POW data
func (b *BlockHeader) bytesWithoutPOW() []byte {
	buff := new(bytes.Buffer)

	// Write version, height of block
	if err := binary.Write(buff, binary.BigEndian, b.Version); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, b.Height); err != nil {
		logrus.Fatal(err)
	}

	// Write prev blockhash
	if len(b.Previous) != BlockHashSize {
		logrus.Fatal(errors.New("invalid previous block hash len"))
	}

	if _, err := buff.Write(b.Previous); err != nil {
		logrus.Fatal(err)
	}

	// Write timestamp
	if err := binary.Write(buff, binary.BigEndian, b.Timestamp.Unix()); err != nil {
		logrus.Fatal(err)
	}

	// Write UTXORoot, RangeProofRoot, KernelRoot
	if len(b.UTXORoot) != BlockHashSize ||
		len(b.RangeProofRoot) != BlockHashSize ||
		len(b.KernelRoot) != BlockHashSize {
		logrus.Fatal(errors.New("invalid UTXORoot/RangeProofRoot/KernelRoot len"))
	}

	if _, err := buff.Write(b.UTXORoot); err != nil {
		logrus.Fatal(err)
	}

	if _, err := buff.Write(b.RangeProofRoot); err != nil {
		logrus.Fatal(err)
	}

	if _, err := buff.Write(b.KernelRoot); err != nil {
		logrus.Fatal(err)
	}

	// Write nonce
	if err := binary.Write(buff, binary.BigEndian, b.Nonce); err != nil {
		logrus.Fatal(err)
	}

	// Write Diff & Total Diff
	if err := binary.Write(buff, binary.BigEndian, uint64(b.Difficulty)); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, uint64(b.TotalDifficulty)); err != nil {
		logrus.Fatal(err)
	}

	return buff.Bytes()
}

func (b *BlockHeader) bytesPOW() []byte {
	buff := new(bytes.Buffer)

	// Write POW
	if uint32(b.POW.ProofSize) != ProofSize || uint32(len(b.POW.Nonces)) != ProofSize {
		logrus.Fatal(errors.New("invalid proof len"))
	}

	for i := 0; i < int(ProofSize); i++ {
		if err := binary.Write(buff, binary.BigEndian, b.POW.Nonces[i]); err != nil {
			logrus.Fatal(err)
		}
	}

	return buff.Bytes()
}

// Bytes implements p2p Message interface
func (b *BlockHeader) Bytes() []byte {
	var buff bytes.Buffer
	buff.Write(b.bytesWithoutPOW())
	buff.Write(b.bytesPOW())

	return buff.Bytes()
}

// Read implements p2p Message interface
func (b *BlockHeader) Read(r io.Reader) error {
	// Read version, height of block
	if err := binary.Read(r, binary.BigEndian, &b.Version); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &b.Height); err != nil {
		return err
	}

	// Read prev blockhash
	b.Previous = make([]byte, BlockHashSize)
	if _, err := io.ReadFull(r, b.Previous); err != nil {
		return err
	}

	// Read timestamp
	var ts int64
	if err := binary.Read(r, binary.BigEndian, &ts); err != nil {
		return err
	}

	b.Timestamp = time.Unix(ts, 0).UTC()

	// Read UTXORoot, RangeProofRoot, KernelRoot
	b.UTXORoot = make([]byte, BlockHashSize)
	if _, err := io.ReadFull(r, b.UTXORoot); err != nil {
		return err
	}

	b.RangeProofRoot = make([]byte, BlockHashSize)
	if _, err := io.ReadFull(r, b.RangeProofRoot); err != nil {
		return err
	}

	b.KernelRoot = make([]byte, BlockHashSize)
	if _, err := io.ReadFull(r, b.KernelRoot); err != nil {
		return err
	}

	// Read nonce
	if err := binary.Read(r, binary.BigEndian, &b.Nonce); err != nil {
		return err
	}

	// Read Diff & Total Diff
	if err := binary.Read(r, binary.BigEndian, &b.Difficulty); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &b.TotalDifficulty); err != nil {
		return err
	}

	// Read POW
	pow := make([]uint32, ProofSize)
	for i := 0; i < int(ProofSize); i++ {
		if err := binary.Read(r, binary.BigEndian, &pow[i]); err != nil {
			return err
		}
	}

	b.POW = NewProof(pow)
	return nil
}

// String implements String() interface
func (p BlockHeader) String() string {
	return fmt.Sprintf("%#v", p)
}