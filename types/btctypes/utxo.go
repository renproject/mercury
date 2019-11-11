package btctypes

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/renproject/mercury/types"
)

const (
	BtcVersion = 2
	BchVersion = 1
	ZecVersion = 4
)

type UTXO interface {
	OutPoint() OutPoint

	SegWit() bool
	TxHash() types.TxHash
	Vout() uint32
	Confirmations() uint64
	Amount() Amount
	ScriptPubKey() []byte
	SigHash(hashType txscript.SigHashType, tx MsgTx, idx int) ([]byte, error)

	Script() []byte
	SetScript(script []byte)
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
	op            OutPoint
	amount        Amount
	scriptPubKey  []byte
	confirmations uint64
	script        []byte
}

func NewUTXO(op OutPoint, amount Amount, scriptPubKey []byte, confirmations uint64, script []byte) UTXO {
	return &utxo{
		op:            op,
		amount:        amount,
		scriptPubKey:  scriptPubKey,
		confirmations: confirmations,
		script:        script,
	}
}

func (u *utxo) Confirmations() uint64 {
	return u.confirmations
}

func (u *utxo) Amount() Amount {
	return u.amount
}

func (u *utxo) ScriptPubKey() []byte {
	return u.scriptPubKey
}

func (u *utxo) TxHash() types.TxHash {
	return u.op.TxHash()
}

func (u *utxo) Vout() uint32 {
	return u.op.Vout()
}

func (u *utxo) OutPoint() OutPoint {
	return u.op
}

func (u *utxo) SigHash(hashType txscript.SigHashType, msgTx MsgTx, idx int) ([]byte, error) {
	scriptPubKey := u.ScriptPubKey()
	switch msgTx := msgTx.(type) {
	case BtcMsgTx:
		if u.script == nil {
			if txscript.IsPayToWitnessPubKeyHash(scriptPubKey) {
				return txscript.CalcWitnessSigHash(scriptPubKey, txscript.NewTxSigHashes(msgTx.MsgTx), hashType, msgTx.MsgTx, idx, int64(u.amount))
			}
			return txscript.CalcSignatureHash(scriptPubKey, hashType, msgTx.MsgTx, idx)
		}
		if txscript.IsPayToWitnessScriptHash(scriptPubKey) {
			return txscript.CalcWitnessSigHash(u.script, txscript.NewTxSigHashes(msgTx.MsgTx), hashType, msgTx.MsgTx, idx, int64(u.amount))
		}
		return txscript.CalcSignatureHash(u.script, hashType, msgTx.MsgTx, idx)
	case *ZecMsgTx:
		if u.script == nil {
			return calcSignatureHash(msgTx.Network, scriptPubKey, hashType, msgTx, idx, u.Amount())
		}
		return calcSignatureHash(msgTx.Network, u.script, hashType, msgTx, idx, u.Amount())
	case BchMsgTx:
		if u.script == nil {
			return calcBip143SignatureHash(scriptPubKey, txscript.NewTxSigHashes(msgTx.MsgTx), hashType, msgTx.MsgTx, idx, u.Amount()), nil
		}
		return calcBip143SignatureHash(u.script, txscript.NewTxSigHashes(msgTx.MsgTx), hashType, msgTx.MsgTx, idx, u.Amount()), nil
	default:
		return nil, fmt.Errorf("unknown msgTx type: %T", msgTx)
	}
}

func (u *utxo) SegWit() bool {
	scriptPubKey := u.ScriptPubKey()
	return txscript.IsPayToWitnessPubKeyHash(scriptPubKey) || txscript.IsPayToWitnessScriptHash(scriptPubKey)
}

func (u *utxo) Script() []byte {
	return u.script
}

func (u *utxo) SetScript(script []byte) {
	u.script = script
}

type OutPoint interface {
	Write(io.Writer) error

	TxHash() types.TxHash
	Vout() uint32
	fmt.Stringer
}

type outPoint struct {
	txHash types.TxHash
	vout   uint32
}

func NewOutPoint(txHash types.TxHash, vout uint32) OutPoint {
	return &outPoint{
		txHash: txHash,
		vout:   vout,
	}
}

func ReadOutPoint(r io.Reader) (OutPoint, error) {
	op := outPoint{}
	if err := binary.Read(r, binary.LittleEndian, &op.txHash); err != nil {
		return &op, err
	}
	if err := binary.Read(r, binary.LittleEndian, &op.vout); err != nil {
		return &op, err
	}
	return &op, nil
}

func (op outPoint) TxHash() types.TxHash {
	return op.txHash
}

func (op outPoint) Vout() uint32 {
	return op.vout
}

func (op outPoint) Write(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, op.txHash); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, op.vout); err != nil {
		return err
	}
	return nil
}

func (op outPoint) String() string {
	return fmt.Sprintf("%s:%d", op.txHash, op.vout)
}

func (op outPoint) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		TxHash types.TxHash `json:"txHash"`
		Vout   uint32       `json:"vout"`
	}{
		TxHash: op.txHash,
		Vout:   op.vout,
	})
}

func (op *outPoint) UnmarshalJSON(data []byte) error {
	tmp := struct {
		TxHash types.TxHash `json:"txHash"`
		Vout   uint32       `json:"vout"`
	}{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	*op = outPoint{txHash: tmp.TxHash, vout: tmp.Vout}
	return nil
}

const SigHashForkID txscript.SigHashType = 0x40

// calcBip143SignatureHash computes the sighash digest of a transaction's
// input using the new, optimized digest calculation algorithm defined
// in BIP0143: https://github.com/bitcoin/bips/blob/master/bip-0143.mediawiki.
// This function makes use of pre-calculated sighash fragments stored within
// the passed HashCache to eliminate duplicate hashing computations when
// calculating the final digest, reducing the complexity from O(N^2) to O(N).
// Additionally, signatures now cover the input value of the referenced unspent
// output. This allows offline, or hardware wallets to compute the exact amount
// being spent, in addition to the final transaction fee. In the case the
// wallet if fed an invalid input amount, the real sighash will differ causing
// the produced signature to be invalid.
func calcBip143SignatureHash(subScript []byte, sigHashes *txscript.TxSigHashes,
	hashType txscript.SigHashType, tx *wire.MsgTx, idx int, amt Amount) []byte {

	// As a sanity check, ensure the passed input index for the transaction
	// is valid.
	if idx > len(tx.TxIn)-1 {
		fmt.Printf("calcBip143SignatureHash error: idx %d but %d txins",
			idx, len(tx.TxIn))
		return nil
	}

	// We'll utilize this buffer throughout to incrementally calculate
	// the signature hash for this transaction.
	var sigHash bytes.Buffer

	// First write out, then encode the transaction's version number.
	var bVersion [4]byte
	binary.LittleEndian.PutUint32(bVersion[:], uint32(tx.Version))
	sigHash.Write(bVersion[:])

	// Next write out the possibly pre-calculated hashes for the sequence
	// numbers of all inputs, and the hashes of the previous outs for all
	// outputs.
	var zeroHash chainhash.Hash

	// If anyone can pay isn't active, then we can use the cached
	// hashPrevOuts, otherwise we just write zeroes for the prev outs.
	if hashType&txscript.SigHashAnyOneCanPay == 0 {
		sigHash.Write(sigHashes.HashPrevOuts[:])
	} else {
		sigHash.Write(zeroHash[:])
	}

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

	// Next, write the outpoint being spent.
	sigHash.Write(tx.TxIn[idx].PreviousOutPoint.Hash[:])
	var bIndex [4]byte
	binary.LittleEndian.PutUint32(bIndex[:], tx.TxIn[idx].PreviousOutPoint.Index)
	sigHash.Write(bIndex[:])

	// For p2wsh outputs, and future outputs, the script code is the
	// original script, with all code separators removed, serialized
	// with a var int length prefix.
	wire.WriteVarBytes(&sigHash, 0, subScript)

	// Next, add the input amount, and sequence number of the input being
	// signed.
	var bAmount [8]byte
	binary.LittleEndian.PutUint64(bAmount[:], uint64(amt))
	sigHash.Write(bAmount[:])
	var bSequence [4]byte
	binary.LittleEndian.PutUint32(bSequence[:], tx.TxIn[idx].Sequence)
	sigHash.Write(bSequence[:])

	// If the current signature mode isn't single, or none, then we can
	// re-use the pre-generated hashoutputs sighash fragment. Otherwise,
	// we'll serialize and add only the target output index to the signature
	// pre-image.
	if hashType&sigHashMask != txscript.SigHashSingle &&
		hashType&sigHashMask != txscript.SigHashNone {
		sigHash.Write(sigHashes.HashOutputs[:])
	} else if hashType&sigHashMask == txscript.SigHashSingle && idx < len(tx.TxOut) {
		var b bytes.Buffer
		wire.WriteTxOut(&b, 0, 0, tx.TxOut[idx])
		sigHash.Write(chainhash.DoubleHashB(b.Bytes()))
	} else {
		sigHash.Write(zeroHash[:])
	}

	// Finally, write out the transaction's locktime, and the sig hash
	// type.
	var bLockTime [4]byte
	binary.LittleEndian.PutUint32(bLockTime[:], tx.LockTime)
	sigHash.Write(bLockTime[:])
	var bHashType [4]byte
	binary.LittleEndian.PutUint32(bHashType[:], uint32(hashType|SigHashForkID))
	sigHash.Write(bHashType[:])

	return chainhash.DoubleHashB(sigHash.Bytes())
}

type upgradeParam struct {
	ActivationHeight uint32
	BranchID         []byte
}
