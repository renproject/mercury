package btctypes

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/codahale/blake2"
)

const (
	sigHashMask                 = 0x1f
	blake2BSigHash              = "ZcashSigHash"
	prevoutsHashPersonalization = "ZcashPrevoutHash"
	sequenceHashPersonalization = "ZcashSequencHash"
	outputsHashPersonalization  = "ZcashOutputsHash"

	versionOverwinter        int32  = 3
	versionOverwinterGroupID uint32 = 0x3C48270
	versionSapling                  = 4
	versionSaplingGroupID           = 0x892f2085
)

// ZecMsgTx zec fork
type ZecMsgTx struct {
	*wire.MsgTx
	Network      ZecNetwork
	ExpiryHeight uint32
}

// witnessMarkerBytes are a pair of bytes specific to the witness encoding. If
// this sequence is encoutered, then it indicates a transaction has iwtness
// data. The first byte is an always 0x00 marker byte, which allows decoders to
// distinguish a serialized transaction with witnesses from a regular (legacy)
// one. The second byte is the Flag field, which at the moment is always 0x01,
// but may be extended in the future to accommodate auxiliary non-committed
// fields.
var witnessMarkerBytes = []byte{0x00, 0x01}

// TxHash generates the Hash for the transaction.
func (msg *ZecMsgTx) TxHash() chainhash.Hash {
	var buf bytes.Buffer
	_ = msg.ZecEncode(&buf, 0, wire.BaseEncoding)
	return chainhash.DoubleHashH(buf.Bytes())
}

// ZecEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the Message interface implementation.
// See Serialize for encoding transactions to be stored to disk, such as in a
// database, as opposed to encoding transactions for the wire.
// msg.Version must be 3 or 4 and may or may not include the overwintered flag
func (msg *ZecMsgTx) ZecEncode(w io.Writer, pver uint32, enc wire.MessageEncoding) error {
	if err := binary.Write(w, binary.LittleEndian, uint32(msg.Version)|(1<<31)); err != nil {
		return err
	}

	var versionGroupID = versionOverwinterGroupID
	if msg.Version == versionSapling {
		versionGroupID = versionSaplingGroupID
	}

	if err := binary.Write(w, binary.LittleEndian, versionGroupID); err != nil {
		return err
	}

	// If the encoding nVersion is set to WitnessEncoding, and the Flags
	// field for the MsgTx aren't 0x00, then this indicates the transaction
	// is to be encoded using the new witness inclusionary structure
	// defined in BIP0144.
	doWitness := enc == wire.WitnessEncoding && msg.HasWitness()
	if doWitness {
		// After the txn's Version field, we include two additional
		// bytes specific to the witness encoding. The first byte is an
		// always 0x00 marker byte, which allows decoders to
		// distinguish a serialized transaction with witnesses from a
		// regular (legacy) one. The second byte is the Flag field,
		// which at the moment is always 0x01, but may be extended in
		// the future to accommodate auxiliary non-committed fields.
		if _, err := w.Write(witnessMarkerBytes); err != nil {
			return err
		}
	}

	count := uint64(len(msg.MsgTx.TxIn))
	if err := writeVarInt(w, pver, count); err != nil {
		return err
	}

	for _, ti := range msg.TxIn {
		if err := writeTxIn(w, pver, msg.Version, ti); err != nil {
			return err
		}
	}

	count = uint64(len(msg.TxOut))
	if err := writeVarInt(w, pver, count); err != nil {
		return err
	}

	for _, to := range msg.TxOut {
		if err := WriteTxOut(w, pver, msg.Version, to); err != nil {
			return err
		}
	}

	// If this transaction is a witness transaction, and the witness
	// encoded is desired, then encode the witness for each of the inputs
	// within the transaction.
	if doWitness {
		for _, ti := range msg.TxIn {
			if err := writeTxWitness(w, pver, msg.Version, ti.Witness); err != nil {
				return err
			}
		}
	}

	if err := binary.Write(w, binary.LittleEndian, msg.LockTime); err != nil {
		return err
	}

	if err := binary.Write(w, binary.LittleEndian, msg.ExpiryHeight); err != nil {
		return err
	}

	if msg.Version == versionSapling {
		// valueBalance
		if err := binary.Write(w, binary.LittleEndian, uint64(0)); err != nil {
			return err
		}

		// nShieldedSpend
		if err := writeVarInt(w, pver, 0); err != nil {
			return err
		}

		// nShieldedOutput
		if err := writeVarInt(w, pver, 0); err != nil {
			return err
		}
	}

	return writeVarInt(w, pver, 0)
}

// WriteTxOut encodes to into the bitcoin protocol encoding for a transaction
// output (TxOut) to w.
//
// NOTE: This function is exported in order to allow txscript to compute the
// new sighashes for witness transactions (BIP0143).
func WriteTxOut(w io.Writer, pver uint32, version int32, to *wire.TxOut) error {
	if err := binary.Write(w, binary.LittleEndian, uint64(to.Value)); err != nil {
		return err
	}
	return writeVarBytes(w, pver, to.PkScript)
}

// writeTxIn encodes ti to the bitcoin protocol encoding for a transaction
// input (TxIn) to w.
func writeTxIn(w io.Writer, pver uint32, version int32, ti *wire.TxIn) error {
	err := writeOutPoint(w, pver, version, &ti.PreviousOutPoint)
	if err != nil {
		return err
	}

	err = writeVarBytes(w, pver, ti.SignatureScript)
	if err != nil {
		return err
	}

	return binary.Write(w, binary.LittleEndian, ti.Sequence)
}

// writeOutPoint encodes op to the bitcoin protocol encoding for an OutPoint
// to w.
func writeOutPoint(w io.Writer, pver uint32, version int32, op *wire.OutPoint) error {
	_, err := w.Write(op.Hash[:])
	if err != nil {
		return err
	}
	return binary.Write(w, binary.LittleEndian, op.Index)
}

// writeTxWitness encodes the bitcoin protocol encoding for a transaction
// input's witness into to w.
func writeTxWitness(w io.Writer, pver uint32, version int32, wit [][]byte) error {
	err := writeVarInt(w, pver, uint64(len(wit)))
	if err != nil {
		return err
	}
	for _, item := range wit {
		err = writeVarBytes(w, pver, item)
		if err != nil {
			return err
		}
	}
	return nil
}

// writeVarInt serializes val to w using a variable number of bytes depending
// on its value.
func writeVarInt(w io.Writer, pver uint32, val uint64) error {
	if val < 0xfd {
		return binary.Write(w, binary.LittleEndian, uint8(val))
	}

	if val <= math.MaxUint16 {
		err := binary.Write(w, binary.LittleEndian, 0xfd)
		if err != nil {
			return err
		}
		return binary.Write(w, binary.LittleEndian, uint16(val))
	}

	if val <= math.MaxUint32 {
		err := binary.Write(w, binary.LittleEndian, 0xfe)
		if err != nil {
			return err
		}
		return binary.Write(w, binary.LittleEndian, uint32(val))
	}

	if err := binary.Write(w, binary.LittleEndian, 0xff); err != nil {
		return err
	}
	return binary.Write(w, binary.LittleEndian, val)
}

// writeVarBytes serializes a variable length byte array to w as a varInt
// containing the number of bytes, followed by the bytes themselves.
func writeVarBytes(w io.Writer, pver uint32, bytes []byte) error {
	slen := uint64(len(bytes))
	err := writeVarInt(w, pver, slen)
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)
	return err
}

// NewTxSigHashes computes, and returns the cached sighashes of the given
// transaction.
func NewTxSigHashes(tx *ZecMsgTx) (h *txscript.TxSigHashes, err error) {
	h = &txscript.TxSigHashes{}

	if h.HashPrevOuts, err = calcHashPrevOuts(tx); err != nil {
		return
	}

	if h.HashSequence, err = calcHashSequence(tx); err != nil {
		return
	}

	if h.HashOutputs, err = calcHashOutputs(tx); err != nil {
		return
	}

	return
}

// calcHashPrevOuts calculates a single hash of all the previous outputs
// (txid:index) referenced within the passed transaction. This calculated hash
// can be re-used when validating all inputs spending segwit outputs, with a
// signature hash type of SigHashAll. This allows validation to re-use previous
// hashing computation, reducing the complexity of validating SigHashAll inputs
// from  O(N^2) to O(N).
func calcHashPrevOuts(tx *ZecMsgTx) (chainhash.Hash, error) {
	var b bytes.Buffer
	for _, in := range tx.TxIn {
		// First write out the 32-byte transaction ID one of whose
		// outputs are being referenced by this input.

		b.Write(in.PreviousOutPoint.Hash[:])

		// Next, we'll encode the index of the referenced output as a
		// little endian integer.
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], in.PreviousOutPoint.Index)
		b.Write(buf[:])
	}

	return blake2bHash(b.Bytes(), []byte(prevoutsHashPersonalization))
}

// calcHashSequence computes an aggregated hash of each of the sequence numbers
// within the inputs of the passed transaction. This single hash can be re-used
// when validating all inputs spending segwit outputs, which include signatures
// using the SigHashAll sighash type. This allows validation to re-use previous
// hashing computation, reducing the complexity of validating SigHashAll inputs
// from O(N^2) to O(N).
func calcHashSequence(tx *ZecMsgTx) (chainhash.Hash, error) {
	var b bytes.Buffer
	for _, in := range tx.TxIn {
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], in.Sequence)
		b.Write(buf[:])
	}

	return blake2bHash(b.Bytes(), []byte(sequenceHashPersonalization))
}

// calcHashOutputs computes a hash digest of all outputs created by the
// transaction encoded using the wire format. This single hash can be re-used
// when validating all inputs spending witness programs, which include
// signatures using the SigHashAll sighash type. This allows computation to be
// cached, reducing the total hashing complexity from O(N^2) to O(N).
func calcHashOutputs(tx *ZecMsgTx) (_ chainhash.Hash, err error) {
	var b bytes.Buffer
	for _, out := range tx.TxOut {
		if err = wire.WriteTxOut(&b, 0, 0, out); err != nil {
			return chainhash.Hash{}, err
		}
	}

	return blake2bHash(b.Bytes(), []byte(outputsHashPersonalization))
}

func calcSignatureHash(
	network ZecNetwork,
	subScript []byte,
	hashType txscript.SigHashType,
	tx *ZecMsgTx,
	idx int,
	amt Amount,
) ([]byte, error) {
	sigHashes, err := NewTxSigHashes(tx)
	if err != nil {
		return nil, err
	}

	// As a sanity check, ensure the passed input index for the transaction
	// is valid.
	if idx > len(tx.TxIn)-1 {
		return nil, fmt.Errorf("blake2bSignatureHash error: idx %d but %d txins", idx, len(tx.TxIn))
	}

	// We'll utilize this buffer throughout to incrementally calculate
	// the signature hash for this transaction.
	var sigHash bytes.Buffer

	// << GetHeader
	// First write out, then encode the transaction's nVersion number. Zcash current nVersion = 3
	var bVersion [4]byte
	binary.LittleEndian.PutUint32(bVersion[:], uint32(tx.Version)|(1<<31))
	sigHash.Write(bVersion[:])

	var versionGroupID = versionOverwinterGroupID
	if tx.Version == versionSapling {
		versionGroupID = versionSaplingGroupID
	}

	// << nVersionGroupId
	// Version group ID
	var nVersion [4]byte
	binary.LittleEndian.PutUint32(nVersion[:], versionGroupID)
	sigHash.Write(nVersion[:])

	// Next write out the possibly pre-calculated hashes for the sequence
	// numbers of all inputs, and the hashes of the previous outs for all
	// outputs.
	var zeroHash chainhash.Hash

	// << hashPrevouts
	// If anyone can pay isn't active, then we can use the cached
	// hashPrevOuts, otherwise we just write zeroes for the prev outs.
	if hashType&txscript.SigHashAnyOneCanPay == 0 {
		sigHash.Write(sigHashes.HashPrevOuts[:])
	} else {
		sigHash.Write(zeroHash[:])
	}

	// << hashSequence
	// If the sighash isn't anyone can pay, single, or none, the use the
	// cached hash sequences, otherwise write all zeroes for the
	// hashSequence.
	if hashType&txscript.SigHashAnyOneCanPay == 0 &&
		hashType&sigHashMask != txscript.SigHashSingle &&
		hashType&sigHashMask != txscript.SigHashNone {
		sigHash.Write(sigHashes.HashSequence[:])
	} else {
		sigHash.Write(zeroHash[:])
	}

	// << hashOutputs
	// If the current signature mode isn't single, or none, then we can
	// re-use the pre-generated hashoutputs sighash fragment. Otherwise,
	// we'll serialize and add only the target output index to the signature
	// pre-image.
	if hashType&sigHashMask != txscript.SigHashSingle && hashType&sigHashMask != txscript.SigHashNone {
		sigHash.Write(sigHashes.HashOutputs[:])
	} else if hashType&sigHashMask == txscript.SigHashSingle && idx < len(tx.TxOut) {
		var (
			b bytes.Buffer
			h chainhash.Hash
		)
		if err := wire.WriteTxOut(&b, 0, 0, tx.TxOut[idx]); err != nil {
			return nil, err
		}

		var err error
		if h, err = blake2bHash(b.Bytes(), []byte(outputsHashPersonalization)); err != nil {
			return nil, err
		}
		sigHash.Write(h.CloneBytes())
	} else {
		sigHash.Write(zeroHash[:])
	}

	// << hashJoinSplits
	sigHash.Write(zeroHash[:])

	// << hashShieldedSpends
	if tx.Version == versionSapling {
		sigHash.Write(zeroHash[:])
	}

	// << hashShieldedOutputs
	if tx.Version == versionSapling {
		sigHash.Write(zeroHash[:])
	}

	// << nLockTime
	var lockTime [4]byte
	binary.LittleEndian.PutUint32(lockTime[:], tx.LockTime)
	sigHash.Write(lockTime[:])

	// << nExpiryHeight
	var expiryTime [4]byte
	binary.LittleEndian.PutUint32(expiryTime[:], tx.ExpiryHeight)
	sigHash.Write(expiryTime[:])

	// << valueBalance
	if tx.Version == versionSapling {
		var valueBalance [8]byte
		binary.LittleEndian.PutUint64(valueBalance[:], 0)
		sigHash.Write(valueBalance[:])
	}

	// << nHashType
	var bHashType [4]byte
	binary.LittleEndian.PutUint32(bHashType[:], uint32(hashType))
	sigHash.Write(bHashType[:])

	if idx != math.MaxUint32 {
		// << prevout
		// Next, write the outpoint being spent.
		sigHash.Write(tx.TxIn[idx].PreviousOutPoint.Hash[:])
		var bIndex [4]byte
		binary.LittleEndian.PutUint32(bIndex[:], tx.TxIn[idx].PreviousOutPoint.Index)
		sigHash.Write(bIndex[:])

		// << scriptCode
		// For p2wsh outputs, and future outputs, the script code is the
		// original script, with all code separators removed, serialized
		// with a var int length prefix.
		// wire.WriteVarBytes(&sigHash, 0, subScript)
		if err := wire.WriteVarBytes(&sigHash, 0, subScript); err != nil {
			return nil, err
		}

		// << amount
		// Next, add the input amount, and sequence number of the input being
		// signed.
		if err := binary.Write(&sigHash, binary.LittleEndian, amt); err != nil {
			return nil, err
		}

		// << nSequence
		var bSequence [4]byte
		binary.LittleEndian.PutUint32(bSequence[:], tx.TxIn[idx].Sequence)
		sigHash.Write(bSequence[:])
	}

	var h chainhash.Hash
	if h, err = blake2bHash(sigHash.Bytes(), sigHashKey(tx.ExpiryHeight, network)); err != nil {
		return nil, err
	}

	return h.CloneBytes(), nil
}

func blake2bHash(data, key []byte) (h chainhash.Hash, err error) {
	bHash := blake2.New(&blake2.Config{
		Size:     32,
		Personal: key,
	})

	if _, err = bHash.Write(data); err != nil {
		return h, err
	}

	err = (&h).SetBytes(bHash.Sum(nil))
	return h, err
}

func sigHashKey(activationHeight uint32, network ZecNetwork) []byte {
	var i int
	upgradeParams := network.upgradeParams
	for i = len(upgradeParams) - 1; i >= 0; i-- {
		if activationHeight >= upgradeParams[i].ActivationHeight {
			break
		}
	}
	return append([]byte(blake2BSigHash), upgradeParams[i].BranchID...)
}
