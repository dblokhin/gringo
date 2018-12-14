// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package consensus

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/dblokhin/gringo/src/secp256k1zkp"
	"github.com/sirupsen/logrus"
	"github.com/yoss22/bulletproofs"
	"golang.org/x/crypto/blake2b"
	"io"
	"sort"
	"time"
)

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

func (f OutputFeatures) String() string {
	switch f {
	case DefaultOutput:
		return ""
	case CoinbaseOutput:
		return "Coinbase"
	}
	return ""
}

// KernelFeatures is options for a kernel's structure or use
type KernelFeatures uint8

const (
	// No flags
	DefaultKernel KernelFeatures = 0
	// Kernel matching a coinbase output
	CoinbaseKernel KernelFeatures = 1 << 0
)

func (f KernelFeatures) String() string {
	switch f {
	case DefaultKernel:
		return ""
	case CoinbaseKernel:
		return "Coinbase"
	}
	return ""
}

// BlockID identify block by Hash or/and Height (if not nill)
type BlockID struct {
	// Block hash, if nil - use the height
	Hash Hash
	// Block height, if nil - use the hash
	Height *uint64
}

// Block of grin blockchain
type Block struct {
	Header BlockHeader

	Inputs  InputList
	Outputs OutputList
	Kernels TxKernelList
}

// Bytes implements p2p Message interface
func (b *Block) Bytes() []byte {
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

	// consensus rule: input, output, kernels MUST BE sorted!
	sort.Sort(b.Inputs)
	sort.Sort(b.Outputs)
	sort.Sort(b.Kernels)

	// Write inputs
	for _, input := range b.Inputs {
		if _, err := buff.Write(input.Bytes()); err != nil {
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

	// Sanity check the lengths.
	if inputs > 1000000 {
		return errors.New("transaction contains too many inputs")
	}
	if outputs > 1000000 {
		return errors.New("transaction contains too many outputs")
	}
	if kernels > 1000000 {
		return errors.New("transaction contains too many kernels")
	}

	// Read inputs
	b.Inputs = make([]Input, inputs)
	for i := uint64(0); i < inputs; i++ {
		if err := b.Inputs[i].Read(r); err != nil {
			return err
		}
	}

	// Read outputs
	b.Outputs = make([]Output, outputs)
	for i := uint64(0); i < outputs; i++ {
		if err := b.Outputs[i].Read(r); err != nil {
			return err
		}
	}

	// Read kernels
	b.Kernels = make([]TxKernel, kernels)
	for i := uint64(0); i < kernels; i++ {
		if err := b.Kernels[i].Read(r); err != nil {
			return err
		}
	}

	return nil
}

// String implements String() interface
func (p Block) String() string {
	return fmt.Sprintf("%#v", p)
}

// Hash returns hash of block
func (b *Block) Hash() Hash {
	return b.Header.Hash()
}

// Validate returns nil if block successfully passed BLOCK-SCOPE consensus rules
func (b *Block) Validate() error {
	logrus.Info("block scope validate")
	/*
		TODO: implement it:

		verify_weight()
		verify_sorted()
		verify_coinbase()
		verify_kernels()

	*/

	// Validate header and proof-of-work.
	if err := b.Header.Validate(); err != nil {
		return err
	}

	// Check that consensus rule MaxBlockCoinbaseOutputs & MaxBlockCoinbaseKernels
	if len(b.Outputs) == 0 || len(b.Kernels) == 0 {
		return errors.New("invalid nocoinbase block")
	}

	// Check sorted inputs, outputs, kernels
	if err := b.verifySorted(); err != nil {
		return err
	}

	if err := b.verifyCoinbase(); err != nil {
		return err
	}

	// Verify all output values are within the correct range.
	if err := b.verifyRangeProofs(); err != nil {
		return err
	}

	if err := b.verifyKernels(); err != nil {
		return err
	}

	return nil
}

func (b *Block) verifyCoinbase() error {
	coinbase := 0

	for _, output := range b.Outputs {
		if output.Features&CoinbaseOutput == CoinbaseOutput {
			coinbase++

			if coinbase > MaxBlockCoinbaseOutputs {
				return errors.New("invalid block with few coinbase outputs")
			}

			// Validate output
			if err := output.Validate(); err != nil {
				return err
			}
		}
	}

	// Check the roots
	// TODO: do that

	return nil
}

func (b *Block) verifyKernels() error {
	coinbase := 0

	for _, kernel := range b.Kernels {
		if kernel.Features&CoinbaseKernel == CoinbaseKernel {
			coinbase++

			if coinbase > MaxBlockCoinbaseKernels {
				return errors.New("invalid block with few coinbase kernels")
			}

			// Validate kernel
			if err := kernel.Validate(); err != nil {
				return err
			}
		}
	}

	// TODO: Verify that the kernel sums are correct.

	// Check the roots
	// TODO: do that

	return nil
}

// verifySorted checks sorted inputs, outputs, kernels
func (b *Block) verifySorted() error {
	if !sort.IsSorted(b.Inputs) {
		return errors.New("block inputs are not sorted")
	}

	if !sort.IsSorted(b.Outputs) {
		return errors.New("block outputs are not sorted")
	}

	if !sort.IsSorted(b.Kernels) {
		return errors.New("block kernels are not sorted")
	}

	return nil
}

// verifyRangeProofs returns nil if all outputs have valid range proofs.
func (b *Block) verifyRangeProofs() error {
	// TODO(yoss22): Batch verify these.
	prover := bulletproofs.NewProver(64)
	for _, output := range b.Outputs {
		if !prover.Verify(output.Commit, output.RangeProof) {
			return fmt.Errorf("proof verification failed for %v %v",
				output.Commit, output.RangeProof)
		}
	}
	return nil
}

// CompactBlock compact version of grin block
// Compact representation of a full block.
// Each input/output/kernel is represented as a short_id.
// A node is reasonably likely to have already seen all tx data (tx broadcast before block)
// and can go request missing tx data from peers if necessary to hydrate a compact block
// into a full block.
type CompactBlock struct {
	// The header with metadata and commitments to the rest of the data
	Header BlockHeader
	// List of full outputs - specifically the coinbase output(s)
	Outputs OutputList
	// List of full kernels - specifically the coinbase kernel(s)
	Kernels TxKernelList
	// List of transaction kernels, excluding those in the full list (short_ids)
	KernelIDs ShortIDList
}

// Bytes implements p2p Message interface
func (b *CompactBlock) Bytes() []byte {
	buff := new(bytes.Buffer)
	if _, err := buff.Write(b.Header.Bytes()); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, uint8(len(b.Outputs))); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, uint8(len(b.Kernels))); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, uint64(len(b.KernelIDs))); err != nil {
		logrus.Fatal(err)
	}

	// consensus rule: input, output, kernels MUST BE sorted!
	sort.Sort(b.Outputs)
	sort.Sort(b.Kernels)
	sort.Sort(b.KernelIDs)

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

	// Write kernels ids
	for _, id := range b.KernelIDs {
		if _, err := buff.Write(id); err != nil {
			logrus.Fatal(err)
		}
	}

	return buff.Bytes()
}

// Type implements p2p Message interface
func (b *CompactBlock) Type() uint8 {
	return MsgTypeCompactBlock
}

// Read implements p2p Message interface
func (b *CompactBlock) Read(r io.Reader) error {
	// Read block header
	if err := b.Header.Read(r); err != nil {
		return err
	}

	// Read counts
	var (
		outputs, kernels uint8
		kernelIDs        uint64
	)

	if err := binary.Read(r, binary.BigEndian, &outputs); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &kernels); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &kernelIDs); err != nil {
		return err
	}

	// Read outputs
	b.Outputs = make(OutputList, outputs)
	for i := uint8(0); i < outputs; i++ {
		if err := b.Outputs[i].Read(r); err != nil {
			return err
		}
	}

	// Read kernels
	b.Kernels = make(TxKernelList, kernels)
	for i := uint8(0); i < kernels; i++ {

		if err := b.Kernels[i].Read(r); err != nil {
			return err
		}
	}

	// Read kernels ids
	b.KernelIDs = make(ShortIDList, kernelIDs)
	for i := uint64(0); i < kernelIDs; i++ {

		shortID := make(ShortID, ShortIDSize)
		if _, err := io.ReadFull(r, shortID); err != nil {
			return err
		}

		b.KernelIDs[i] = shortID
	}

	return nil
}

// String implements String() interface
func (p CompactBlock) String() string {
	return fmt.Sprintf("%#v", p)
}

// Hash returns hash of block
func (b *CompactBlock) Hash() Hash {
	return b.Header.Hash()
}

type BlockList []Block

type Input struct {
	Features OutputFeatures
	Commit   secp256k1zkp.Commitment
}

func (input *Input) Bytes() []byte {
	buff := new(bytes.Buffer)

	if err := binary.Write(buff, binary.BigEndian, uint8(input.Features)); err != nil {
		logrus.Fatal(err)
	}

	if _, err := buff.Write(input.Commit); err != nil {
		logrus.Fatal(err)
	}

	return buff.Bytes()
}

func (input *Input) Read(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &input.Features); err != nil {
		return err
	}

	commitment := make([]byte, secp256k1zkp.PedersenCommitmentSize)
	if _, err := io.ReadFull(r, commitment); err != nil {
		return err
	}

	input.Commit = commitment

	return nil
}

// Hash returns a hash of the serialised input.
func (input *Input) Hash() []byte {
	hashed := blake2b.Sum256(input.Bytes())
	return hashed[:]
}

// InputList sortable list of inputs
type InputList []Input

func (m InputList) Len() int {
	return len(m)
}

// Less is used to order inputs by their hash.
func (m InputList) Less(i, j int) bool {
	return bytes.Compare(m[i].Hash(), m[j].Hash()) < 0
}

func (m InputList) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
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
	Commit *bulletproofs.Point
	// A proof that the commitment is in the right range
	RangeProof bulletproofs.BulletProof
}

func (o *Output) BytesWithoutProof() []byte {
	buff := new(bytes.Buffer)

	// Write features
	if err := binary.Write(buff, binary.BigEndian, uint8(o.Features)); err != nil {
		logrus.Fatal(err)
	}

	if _, err := buff.Write(o.Commit.Bytes()); err != nil {
		logrus.Fatal(err)
	}

	return buff.Bytes()
}

// Bytes implements p2p Message interface
func (o *Output) Bytes() []byte {
	buff := new(bytes.Buffer)

	if _, err := buff.Write(o.BytesWithoutProof()); err != nil {
		logrus.Fatal(err)
	}

	proof := o.RangeProof.Bytes()

	if err := binary.Write(buff, binary.BigEndian, uint64(len(proof))); err != nil {
		logrus.Fatal(err)
	}

	if _, err := buff.Write(proof); err != nil {
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
	o.Commit = new(bulletproofs.Point)
	if err := o.Commit.Read(r); err != nil {
		return err
	}

	// Read range proof
	var proofLen uint64 // tha max is MaxProofSize (5134), but in message field it is uint64
	if err := binary.Read(r, binary.BigEndian, &proofLen); err != nil {
		return err
	}

	if proofLen > uint64(secp256k1zkp.MaxProofSize) {
		return fmt.Errorf("invalid range proof length: %d", proofLen)
	}

	proof := new(bulletproofs.BulletProof)
	err := proof.Read(io.LimitReader(r, int64(proofLen)))
	if err != nil {
		return errors.New("failed to deserialize range proof")
	}
	o.RangeProof = *proof

	return nil
}

// Validate returns nil if output successfully passed consensus rules
func (o *Output) Validate() error {
	return nil
}

// String implements String() interface
func (p Output) String() string {
	return fmt.Sprintf("%#v", p)
}

// Hash returns a hash of the serialised output.
func (o *Output) Hash() []byte {
	hashed := blake2b.Sum256(o.BytesWithoutProof())
	return hashed[:]
}

// OutputList sortable list of outputs
type OutputList []Output

func (m OutputList) Len() int {
	return len(m)
}

// Less is used to order outputs by their hash.
func (m OutputList) Less(i, j int) bool {
	return bytes.Compare(m[i].Hash(), m[j].Hash()) < 0
}

func (m OutputList) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
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
	Excess bulletproofs.Point
	// The signature proving the excess is a valid public key, which signs
	// the transaction fee.
	ExcessSig [64]byte
}

// Hash returns a hash of the serialised kernel.
func (k *TxKernel) Hash() []byte {
	hashed := blake2b.Sum256(k.Bytes())
	return hashed[:]
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
	if _, err := buff.Write(k.Excess.Bytes()); err != nil {
		logrus.Fatal(err)
	}

	// Write ExcessSig
	if _, err := buff.Write(k.ExcessSig[:]); err != nil {
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
	if err := k.Excess.Read(r); err != nil {
		return err
	}

	if _, err := io.ReadFull(r, k.ExcessSig[:]); err != nil {
		return err
	}

	return nil
}

var ErrInvalidSignature = errors.New("signature isn't valid")

// Validate returns nil if kernel successfully passed consensus rules.
func (o *TxKernel) Validate() error {
	// The spender signs the fee and lock height using the private key for P. If
	// the signature verifies then we know that there is no residue on G (i.e.
	// that no value is created) and that the spender is in possession of the
	// inputs.
	msg := secp256k1zkp.ComputeMessage(o.Fee, o.LockHeight)
	signature := secp256k1zkp.DecodeSignature(o.ExcessSig)

	// Excess is a Pedersen commitment to the value zero: P = γ*H + 0*G
	P := o.Excess

	if !secp256k1zkp.VerifySignature(P, msg, signature) {
		return ErrInvalidSignature
	}

	return nil
}

// String implements String() interface
func (p TxKernel) String() string {
	return fmt.Sprintf("%#v", p)
}

// TxKernelList sortable list of kernels
type TxKernelList []TxKernel

func (m TxKernelList) Len() int {
	return len(m)
}

// Less is used to order kernels by their hash.
func (m TxKernelList) Less(i, j int) bool {
	return bytes.Compare(m[i].Hash(), m[j].Hash()) < 0
}

func (m TxKernelList) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

// BlockHeader header of the grin-blocks
type BlockHeader struct {
	// Version of the block
	Version uint16
	// Height of this block since the genesis block (height 0)
	Height uint64
	// Hash of the block previous to this in the chain
	Previous Hash
	// Root hash of the previous header MMR.
	PreviousRoot Hash
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
	// Total accumulated sum of kernel offsets since genesis block.
	TotalKernelOffset Hash
	// Total accumulated sum of kernel commitments since genesis block.
	// Should always equal the UTXO commitment sum minus supply.
	TotalKernelSum secp256k1zkp.Commitment
	// Total size of the output MMR after applying this block
	OutputMmrSize uint64
	// Total size of the kernel MMR after applying this block
	KernelMmrSize uint64
	// Proof of work.
	POW Proof
	// TODO: Remove or calculate this correctly.
	// Difficulty used to mine the block.
	Difficulty Difficulty
	// Total accumulated difficulty since genesis block
	TotalDifficulty Difficulty
	// Difficulty scaling factor between the different proofs of work
	ScalingDifficulty uint32
}

// Hash is a hash based on the blocks proof of work.
func (b *BlockHeader) Hash() Hash {
	hash := blake2b.Sum256(b.POW.ProofBytes())

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

	// Write timestamp
	if err := binary.Write(buff, binary.BigEndian, b.Timestamp.Unix()); err != nil {
		logrus.Fatal(err)
	}

	// Write prev blockhash
	if len(b.Previous) != BlockHashSize {
		logrus.Fatal(errors.New("invalid previous block hash len"))
	}

	if _, err := buff.Write(b.Previous); err != nil {
		logrus.Fatal(err)
	}

	if len(b.PreviousRoot) != BlockHashSize {
		logrus.Fatal(errors.New("invalid previous root hash len"))
	}

	if _, err := buff.Write(b.PreviousRoot); err != nil {
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

	if _, err := buff.Write(b.TotalKernelOffset); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, b.OutputMmrSize); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, b.KernelMmrSize); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, uint64(b.TotalDifficulty)); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, b.ScalingDifficulty); err != nil {
		logrus.Fatal(err)
	}

	// Write nonce
	if err := binary.Write(buff, binary.BigEndian, b.Nonce); err != nil {
		logrus.Fatal(err)
	}

	return buff.Bytes()
}

func (b *BlockHeader) bytesPOW() []byte {
	return b.POW.Bytes()
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

	// Read timestamp
	var ts int64
	if err := binary.Read(r, binary.BigEndian, &ts); err != nil {
		return err
	}

	// FIXME: Check timestamp is in correct range.
	b.Timestamp = time.Unix(ts, 0).UTC()

	// Read prev blockhash
	b.Previous = make([]byte, BlockHashSize)
	if _, err := io.ReadFull(r, b.Previous); err != nil {
		return err
	}

	b.PreviousRoot = make([]byte, BlockHashSize)
	if _, err := io.ReadFull(r, b.PreviousRoot); err != nil {
		return err
	}

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

	b.TotalKernelOffset = make([]byte, secp256k1zkp.SecretKeySize)
	if _, err := io.ReadFull(r, b.TotalKernelOffset); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &b.OutputMmrSize); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &b.KernelMmrSize); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &b.TotalDifficulty); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &b.ScalingDifficulty); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &b.Nonce); err != nil {
		return err
	}

	if err := b.POW.Read(r); err != nil {
		return err
	}

	return nil
}

// Validate returns nil if header successfully passed consensus rules
func (b *BlockHeader) Validate() error {
	logrus.Info("block header validate")

	// Check block header version
	if !ValidateBlockVersion(b.Height, b.Version) {
		return fmt.Errorf("invalid block version %d on height %d, maybe update Gringo?", b.Version, b.Height)
	}

	// refuse blocks more than 12 blocks intervals in future (as in bitcoin)
	if b.Timestamp.Sub(time.Now().UTC()) > time.Second*12*BlockTimeSec {
		return fmt.Errorf("invalid block time (%s)", b.Timestamp)
	}

	// TODO: Check difficulty.

	// Check POW
	isPrimaryPow := b.POW.EdgeBits != SecondPowEdgeBits

	// Either the size shift must be a valid primary POW (greater than the
	// minimum size shift) or equal to the secondary POW size shift.
	if b.POW.EdgeBits < DefaultMinEdgeBits && isPrimaryPow {
		return fmt.Errorf("cuckoo size too small: %d", b.POW.EdgeBits)
	}

	// The primary POW must have a scaling factor of 1.
	if isPrimaryPow && b.ScalingDifficulty != 1 {
		return fmt.Errorf("invalid scaling difficulty: %d", b.ScalingDifficulty)
	}

	if err := b.POW.Validate(b, b.POW.EdgeBits); err != nil {
		return err
	}

	return nil
}

// ValidateBlockVersion helper for validation block header version
func ValidateBlockVersion(height uint64, version uint16) bool {
	if height < HardForkV2Height {
		return version == 1
	} else if height < HardForkInterval {
		return version == 2
	} else if height < 2*HardForkInterval {
		return version == 3
	} else {
		return false
	}
}

// String implements String() interface
func (p BlockHeader) String() string {
	return fmt.Sprintf("%#v", p)
}
