package btctypes

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"io"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/renproject/mercury/types"
)

type BtcTx interface {
	types.Tx
	UTXOs() UTXOs
	Recipients() Recipients
	OutputUTXO(address Address) UTXO
}

type tx struct {
	outputUTXOs map[string]UTXO
	network     Network
	sigHashes   []types.SignatureHash
	utxos       UTXOs
	recipients  Recipients
	tx          MsgTx
	signed      bool
}

func NewUnsignedTx(network Network, utxos UTXOs, recipients Recipients) (BtcTx, error) {
	outputUTXOs := map[string]UTXO{}
	msgTx := NewMsgTx(network)
	for _, utxo := range utxos {
		hash, err := chainhash.NewHashFromStr(string(utxo.TxHash()))
		if err != nil {
			return nil, err
		}
		msgTx.AddTxIn(wire.NewTxIn(wire.NewOutPoint(hash, utxo.Vout()), nil, nil))
	}
	for i, recipient := range recipients {
		script, err := PayToAddrScript(recipient.Address, network)
		if err != nil {
			return nil, err
		}
		msgTx.AddTxOut(wire.NewTxOut(int64(recipient.Amount), script))
		outputUTXOs[recipient.Address.EncodeAddress()] = NewUTXO(NewOutPoint("", uint32(i)), recipient.Amount, script, 0, nil)
	}
	t := tx{
		outputUTXOs: outputUTXOs,
		network:     network,
		sigHashes:   []types.SignatureHash{},
		tx:          msgTx,
		utxos:       utxos,
		signed:      false,
	}
	for i, utxo := range utxos {
		sigHash, err := utxo.SigHash(txscript.SigHashAll, msgTx, i)
		if err != nil {
			return nil, err
		}
		t.sigHashes = append(t.sigHashes, sigHash)
	}
	return &t, nil
}

// SignatureHashes returns a list of signature hashes need to be signed.
func (t *tx) SignatureHashes() []types.SignatureHash {
	return t.sigHashes
}

func (t *tx) Sign(key *ecdsa.PrivateKey) (err error) {
	subScripts := t.SignatureHashes()
	sigs := make([]*btcec.Signature, len(subScripts))
	for i, subScript := range subScripts {
		sigs[i], err = (*btcec.PrivateKey)(key).Sign(subScript)
		if err != nil {
			return err
		}
	}
	return t.InjectSignatures(sigs, key.PublicKey)
}

func (t *tx) IsSigned() bool {
	return t.signed
}

// OutPoint returns the OutPoint that can is funding the given address, returns
// nil if the address is not being funded.
func (t *tx) OutputUTXO(address Address) UTXO {
	if !t.signed {
		panic("OutPoint should only be called after signing the transaction")
	}
	utxo, ok := t.outputUTXOs[address.EncodeAddress()]
	if !ok {
		return nil
	}
	return NewUTXO(NewOutPoint(t.Hash(), utxo.Vout()), utxo.Amount(), utxo.ScriptPubKey(), 0, nil)
}

// Serialize returns the serialized tx in bytes.
func (t *tx) Serialize() ([]byte, error) {
	// Pre-condition checks
	if t.tx == nil {
		panic("pre-condition violation: cannot serialize nil transaction")
	}

	buf := bytes.NewBuffer([]byte{})
	if err := t.tx.Serialize(buf); err != nil {
		return nil, err
	}
	bufBytes := buf.Bytes()

	// Post-condition checks
	if len(bufBytes) <= 0 {
		panic(fmt.Errorf("post-condition violation: serialized transaction len=%v is invalid", len(bufBytes)))
	}
	return bufBytes, nil
}

func (t *tx) Hash() types.TxHash {
	return types.TxHash(t.tx.TxHash().String())
}

// InjectSignatures injects the signed signatureHashes into the Tx. You cannot use the USTX anymore. scriptData is
// additional data to be appended to the signature script, it can be nil or an empty byte array.
func (t *tx) InjectSignatures(sigs []*btcec.Signature, pubKey ecdsa.PublicKey) error {
	// Pre-condition checks
	if t.IsSigned() {
		panic("pre-condition violation: cannot inject signatures into signed transaction")
	}
	if t.tx == nil {
		panic("pre-condition violation: cannot inject signatures into nil transaction")
	}
	if len(sigs) <= 0 {
		panic("pre-condition violation: cannot inject empty signatures")
	}
	if len(sigs) != t.tx.InCount() {
		panic(fmt.Errorf("pre-condition violation: expected signature len=%v to equal transaction input len=%v", len(sigs), t.tx.InCount()))
	}
	serializedPubKey := SerializePublicKey(pubKey)
	if len(serializedPubKey) <= 0 {
		panic("pre-condition violation: cannot inject signatures with empty pubkey")
	}

	for i, sig := range sigs {
		if !t.utxos[i].SegWit() {
			builder := txscript.NewScriptBuilder()
			builder.AddData(t.tx.SigBytes(sig, txscript.SigHashAll))
			builder.AddData(serializedPubKey)
			if script := t.utxos[i].Script(); script != nil {
				builder.AddData(script)
			}
			sigScript, err := builder.Script()
			if err != nil {
				return err
			}
			t.tx.AddSigScript(i, sigScript)
		} else {
			if script := t.utxos[i].Script(); script != nil {
				t.tx.AddSegWit(i, append(sig.Serialize(), byte(txscript.SigHashAll)), serializedPubKey, script)
			} else {
				t.tx.AddSegWit(i, append(sig.Serialize(), byte(txscript.SigHashAll)), serializedPubKey)
			}
		}
	}
	t.signed = true
	return nil
}

func (t *tx) UTXOs() UTXOs {
	return t.utxos
}

func (t *tx) Recipients() Recipients {
	return t.recipients
}

type MsgTx interface {
	Serialize(buffer io.Writer) error
	TxHash() chainhash.Hash
	InCount() int
	AddTxIn(txIn *wire.TxIn)
	AddTxOut(txOut *wire.TxOut)
	AddSigScript(i int, sigScript []byte)
	AddSegWit(i int, witness ...[]byte)
	SigBytes(sig *btcec.Signature, hashType txscript.SigHashType) []byte
}

func NewMsgTx(network Network) MsgTx {
	switch network.Chain() {
	case types.Bitcoin:
		return NewBtcMsgTx(wire.NewMsgTx(BtcVersion))
	case types.ZCash:
		return NewZecMsgTx(network.(ZecNetwork), wire.NewMsgTx(ZecVersion), ZecExpiryHeight)
	case types.BitcoinCash:
		return NewBchMsgTx(wire.NewMsgTx(BchVersion))
	default:
		panic(types.ErrUnknownChain)
	}
}

type BtcMsgTx struct {
	*wire.MsgTx
}

func NewBtcMsgTx(msgTx *wire.MsgTx) BtcMsgTx {
	return BtcMsgTx{msgTx}
}

func (msgTx BtcMsgTx) InCount() int {
	return len(msgTx.TxIn)
}

func (msgTx BtcMsgTx) AddSigScript(i int, sigScript []byte) {
	msgTx.TxIn[i].SignatureScript = sigScript
}

func (msgTx BtcMsgTx) AddSegWit(i int, witness ...[]byte) {
	msgTx.TxIn[i].Witness = wire.TxWitness(witness)
}

func (BtcMsgTx) SigBytes(sig *btcec.Signature, hashType txscript.SigHashType) []byte {
	return append(sig.Serialize(), byte(hashType))
}

func (msgTx *ZecMsgTx) Serialize(buf io.Writer) error {
	return msgTx.ZecEncode(buf, 0, wire.BaseEncoding)
}

func (msgTx *ZecMsgTx) InCount() int {
	return len(msgTx.TxIn)
}

func (msgTx *ZecMsgTx) AddSigScript(i int, sigScript []byte) {
	msgTx.TxIn[i].SignatureScript = sigScript
}

func (msgTx *ZecMsgTx) AddSegWit(i int, witness ...[]byte) {
	panic(ErrDoesNotSupportSegWit)
}

func (*ZecMsgTx) SigBytes(sig *btcec.Signature, hashType txscript.SigHashType) []byte {
	return append(sig.Serialize(), byte(hashType))
}

func NewZecMsgTx(network ZecNetwork, msgTx *wire.MsgTx, expiryHeight uint32) *ZecMsgTx {
	return &ZecMsgTx{
		Network:      network,
		MsgTx:        msgTx,
		ExpiryHeight: expiryHeight,
	}
}

type BchMsgTx struct {
	*wire.MsgTx
}

func (msgTx BchMsgTx) InCount() int {
	return len(msgTx.TxIn)
}

func (msgTx BchMsgTx) AddSigScript(i int, sigScript []byte) {
	msgTx.TxIn[i].SignatureScript = sigScript
}

func (msgTx BchMsgTx) AddSegWit(i int, witness ...[]byte) {
	panic("BitcoinCash does not support SegWit")
}

func (BchMsgTx) SigBytes(sig *btcec.Signature, hashType txscript.SigHashType) []byte {
	return append(sig.Serialize(), byte(hashType|SigHashForkID))
}

func NewBchMsgTx(msgTx *wire.MsgTx) BchMsgTx {
	return BchMsgTx{msgTx}
}

func EstimateTxSize(numUTXOs, numRecipients int) int {
	return 146*numUTXOs + 33*numRecipients + 10
}

var ErrDoesNotSupportSegWit = fmt.Errorf("this blockchain does not support SegWit")
