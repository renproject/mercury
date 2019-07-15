package btctx

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/renproject/mercury/types/btctypes/btcaddress"
	"github.com/renproject/mercury/types/btctypes/btcutxo"
)

type btcTx struct {
	network   btctypes.Network
	sigHashes []types.SignatureHash
	utxos     btcutxo.UTXOs
	tx        *wire.MsgTx
	signed    bool
}

func NewUnsignedBtcTx(network btctypes.Network, utxos btcutxo.UTXOs, msgTx *wire.MsgTx) (BtcTx, error) {
	t := btcTx{
		network:   network,
		sigHashes: []types.SignatureHash{},
		tx:        msgTx,
		utxos:     utxos,
		signed:    false,
	}

	buf := new(bytes.Buffer)
	if err := msgTx.Serialize(buf); err != nil {
		return nil, err
	}

	for i, utxo := range utxos {
		sigHash, err := utxo.SigHash(txscript.SigHashAll, btcutxo.NewBtcMsgTx(msgTx), i)
		if err != nil {
			return nil, err
		}
		t.sigHashes = append(t.sigHashes, sigHash)
	}
	return &t, nil
}

// SignatureHashes returns a list of signature hashes need to be signed.
func (t *btcTx) SignatureHashes() []types.SignatureHash {
	return t.sigHashes
}

func (t *btcTx) Sign(key *ecdsa.PrivateKey) (err error) {
	subScripts := t.SignatureHashes()
	sigs := make([]*btcec.Signature, len(subScripts))
	for i, subScript := range subScripts {
		sigs[i], err = (*btcec.PrivateKey)(key).Sign(subScript)
		if err != nil {
			return err
		}
	}
	serializedPK := btcaddress.SerializePublicKey(&key.PublicKey, t.network)
	return t.InjectSignatures(sigs, serializedPK)
}

func (t *btcTx) IsSigned() bool {
	return t.signed
}

// Serialize returns the serialized tx in bytes.
func (t *btcTx) Serialize() ([]byte, error) {
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

func (t *btcTx) Hash() types.TxHash {
	return types.TxHash(t.tx.TxHash().String())
}

// InjectSignatures injects the signed signatureHashes into the Tx. You cannot use the USTX anymore. scriptData is
// additional data to be appended to the signature script, it can be nil or an empty byte array.
func (t *btcTx) InjectSignatures(sigs []*btcec.Signature, serializedPubKey []byte) error {
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
	if len(sigs) != len(t.tx.TxIn) {
		panic(fmt.Errorf("pre-condition violation: expected signature len=%v to equal transaction input len=%v", len(sigs), len(t.tx.TxIn)))
	}
	if len(serializedPubKey) <= 0 {
		panic("pre-condition violation: cannot inject signatures with empty pubkey")
	}

	for i, sig := range sigs {
		builder := txscript.NewScriptBuilder()
		builder.AddData(append(sig.Serialize(), byte(txscript.SigHashAll)))
		builder.AddData(serializedPubKey)
		t.utxos[i].AddData(builder)
		sigScript, err := builder.Script()
		if err != nil {
			return err
		}
		t.tx.TxIn[i].SignatureScript = sigScript
	}
	t.signed = true
	return nil
}

func (t *btcTx) UTXOs() btcutxo.UTXOs {
	return t.utxos
}
