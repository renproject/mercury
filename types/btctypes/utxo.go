package btctypes

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/codahale/blake2"
	"github.com/iqoption/zecutil"
	"github.com/renproject/mercury/types"
)

const (
	BtcVersion = 2

	ZecExpiryHeight = uint32(10000000)
	ZecVersion      = 4
)

type UTXO interface {
	OutPoint

	Confirmations() uint64
	Amount() Amount
	ScriptPubKey() []byte
	SigHash(hashType txscript.SigHashType, tx MsgTx, idx int) ([]byte, error)
	AddData(builder *txscript.ScriptBuilder)
}

type UTXOs []UTXO

func (utxos UTXOs) Sum() Amount {
	total := Amount(0)
	for _, utxo := range utxos {
		total += utxo.Amount()
	}
	return total
}

func (utxos UTXOs) Filter(confs uint64) UTXOs {
	newList := UTXOs{}
	for _, utxo := range utxos {
		if utxo.Confirmations() >= confs {
			newList = append(newList, utxo)
		}
	}
	return newList
}

type utxo struct {
	OutPoint

	amount          Amount
	scriptPubKey    []byte
	confirmations   uint64
	Script          []byte
	UpdateSigScript func(builder *txscript.ScriptBuilder)
}

func (u utxo) Confirmations() uint64 {
	return u.confirmations
}

func (u utxo) Amount() Amount {
	return u.amount
}

func (u utxo) ScriptPubKey() []byte {
	return u.scriptPubKey
}

func (u utxo) SigHash(hashType txscript.SigHashType, msgTx MsgTx, idx int) ([]byte, error) {
	switch msgTx := msgTx.(type) {
	case BtcMsgTx:
		if u.Script == nil {
			return txscript.CalcSignatureHash(u.ScriptPubKey(), hashType, msgTx.MsgTx, idx)
		}
		return txscript.CalcSignatureHash(u.Script, hashType, msgTx.MsgTx, idx)
	case ZecMsgTx:
		if u.Script == nil {
			return calcSignatureHash(u.ScriptPubKey(), hashType, msgTx.MsgTx, idx, u.Amount())
		}
		return calcSignatureHash(u.Script, hashType, msgTx.MsgTx, idx, u.Amount())
	default:
		return nil, fmt.Errorf("unknown msgTx type: %T", msgTx)
	}
}

func (u utxo) AddData(builder *txscript.ScriptBuilder) {
	if u.UpdateSigScript != nil {
		u.UpdateSigScript(builder)
	}
}

func NewUTXO(op OutPoint, amount Amount, scriptPubKey []byte, confirmations uint64, script []byte, updateSigScript func(builder *txscript.ScriptBuilder)) UTXO {
	return utxo{
		OutPoint:        op,
		amount:          amount,
		scriptPubKey:    scriptPubKey,
		confirmations:   confirmations,
		UpdateSigScript: updateSigScript,
		Script:          script,
	}
}

type OutPoint interface {
	TxHash() types.TxHash
	Vout() uint32
}

type outPoint struct {
	txHash types.TxHash
	vout   uint32
}

func NewOutPoint(txHash types.TxHash, vout uint32) OutPoint {
	return outPoint{
		txHash: txHash,
		vout:   vout,
	}
}

func (op outPoint) TxHash() types.TxHash {
	return op.txHash
}

func (op outPoint) Vout() uint32 {
	return op.vout
}

type upgradeParam struct {
	ActivationHeight uint32
	BranchID         []byte
}

const (
	sigHashMask                = 0x1f
	blake2BSigHash             = "ZcashSigHash"
	outputsHashPersonalization = "ZcashOutputsHash"

	versionOverwinter        int32  = 3
	versionOverwinterGroupID uint32 = 0x3C48270
	versionSapling                  = 4
	versionSaplingGroupID           = 0x892f2085
)

var upgradeParams = []upgradeParam{
	{0, []byte{0x00, 0x00, 0x00, 0x00}},
	{207500, []byte{0x19, 0x1B, 0xA8, 0x5B}},
	{280000, []byte{0xBB, 0x09, 0xB8, 0x76}},
}

func calcSignatureHash(
	subScript []byte,
	hashType txscript.SigHashType,
	tx *zecutil.MsgTx,
	idx int,
	amt Amount,
) ([]byte, error) {
	sigHashes, err := zecutil.NewTxSigHashes(tx)
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
	if h, err = blake2bHash(sigHash.Bytes(), sigHashKey(tx.ExpiryHeight)); err != nil {
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

func sigHashKey(activationHeight uint32) []byte {
	var i int
	for i = len(upgradeParams) - 1; i >= 0; i-- {
		if activationHeight >= upgradeParams[i].ActivationHeight {
			break
		}
	}

	return append([]byte(blake2BSigHash), upgradeParams[i].BranchID...)
}