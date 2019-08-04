package bnctypes

import (
	"crypto/ecdsa"

	"github.com/binance-chain/go-sdk/types/tx"
	"github.com/btcsuite/btcd/btcec"
	"github.com/renproject/mercury/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

type BNCTx interface {
	types.Tx
}

type bncTx struct {
	msg    tx.StdSignMsg
	tx     tx.StdTx
	signed bool
}

func NewTx(msg tx.StdSignMsg) BNCTx {
	return &bncTx{
		msg: msg,
	}
}

func (bncTx *bncTx) SignatureHashes() []types.SignatureHash {
	return []types.SignatureHash{crypto.Sha256(bncTx.msg.Bytes())}
}

func (bncTx *bncTx) Sign(key *ecdsa.PrivateKey) (err error) {
	hashes := bncTx.SignatureHashes()
	sigs := make([]*btcec.Signature, len(hashes))
	privKey := (*btcec.PrivateKey)(key)
	for i, hash := range hashes {
		sigs[i], err = privKey.Sign(hash)
		if err != nil {
			return err
		}
	}
	return bncTx.InjectSignatures(sigs, privKey.PubKey().SerializeCompressed())
}

func (bncTx *bncTx) IsSigned() bool {
	return bncTx.signed
}

func (bncTx *bncTx) Serialize() ([]byte, error) {
	return tx.Cdc.MarshalBinaryLengthPrefixed(&bncTx.tx)
}

func (bncTx *bncTx) Hash() types.TxHash {
	panic("unimplemented")
}

func (bncTx *bncTx) InjectSignatures(sigs []*btcec.Signature, serializedPubKey []byte) error {
	var pubkeyBytes secp256k1.PubKeySecp256k1
	copy(pubkeyBytes[:], serializedPubKey)
	bncTx.tx = tx.NewStdTx(bncTx.msg.Msgs, []tx.StdSignature{tx.StdSignature{
		AccountNumber: bncTx.msg.AccountNumber,
		Sequence:      bncTx.msg.Sequence,
		PubKey:        pubkeyBytes,
		Signature:     sigs[0].Serialize(),
	}}, bncTx.msg.Memo, bncTx.msg.Source, bncTx.msg.Data)
	bncTx.signed = true
	return nil
}
