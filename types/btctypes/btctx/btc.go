package btctx

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/txscript"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/renproject/mercury/types/btctypes/btcaddress"
	"github.com/renproject/mercury/types/btctypes/btcutxo"
)

type tx struct {
	outputUTXOs map[btcaddress.Address]btcutxo.UTXO
	network     btctypes.Network
	sigHashes   []types.SignatureHash
	utxos       btcutxo.UTXOs
	tx          btcutxo.MsgTx
	signed      bool
}

func NewUnsignedTx(network btctypes.Network, utxos btcutxo.UTXOs, msgTx btcutxo.MsgTx, outputUTXOs map[btcaddress.Address]btcutxo.UTXO) (BtcTx, error) {
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
	serializedPK := btcaddress.SerializePublicKey(key.PublicKey, t.network)
	return t.InjectSignatures(sigs, serializedPK)
}

func (t *tx) IsSigned() bool {
	return t.signed
}

// OutPoint returns the OutPoint that can is funding the given address, returns
// nil if the address is not being funded.
func (t *tx) OutputUTXO(address btcaddress.Address) btcutxo.UTXO {
	if !t.signed {
		panic("OutPoint should only be called after signing the transaction")
	}
	utxo, ok := t.outputUTXOs[address]
	if !ok {
		return nil
	}
	return btcutxo.NewStandardUTXO(t.network.Chain(), t.Hash(), utxo.Amount(), utxo.ScriptPubKey(), utxo.Vout(), 0)
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
func (t *tx) InjectSignatures(sigs []*btcec.Signature, serializedPubKey []byte) error {
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
		t.tx.AddSigScript(i, sigScript)
	}
	t.signed = true
	return nil
}

func (t *tx) UTXOs() btcutxo.UTXOs {
	return t.utxos
}
