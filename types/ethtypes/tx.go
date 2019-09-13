package ethtypes

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	mTypes "github.com/renproject/mercury/types"
)

type ethTx struct {
	chainID *big.Int
	tx      *types.Transaction
	signed  bool
}

// SignatureHashes() []SignatureHash
// InjectSignatures(sigs []*btcec.Signature, pubKey ecdsa.PublicKey) error

func (tx *ethTx) Hash() mTypes.TxHash {
	return mTypes.TxHash(tx.tx.Hash().String())
}

func (tx *ethTx) IsSigned() bool {
	return tx.signed
}

func (tx *ethTx) SignatureHashes() []mTypes.SignatureHash {
	signer := types.NewEIP155Signer(tx.chainID)
	return []mTypes.SignatureHash{signer.Hash(tx.tx).Bytes()}
}

func (tx *ethTx) InjectSigs(sigs [][]byte, _ ecdsa.PublicKey) error {
	signer := types.NewEIP155Signer(tx.chainID)
	stx, err := tx.tx.WithSignature(signer, sigs[0])
	if err != nil {
		return err
	}
	tx.tx = stx
	tx.signed = true
	return nil
}

func (tx *ethTx) Serialize() ([]byte, error) {
	return tx.tx.MarshalJSON()
}

func (tx *ethTx) Sign(key *ecdsa.PrivateKey) error {
	// Pre-condition checks
	if tx.IsSigned() {
		panic("pre-condition violation: cannot sign already signed transaction")
	}
	hashes := tx.SignatureHashes()
	sig, err := crypto.Sign(hashes[0], key)
	if err != nil {
		return err
	}
	return tx.InjectSigs([][]byte{sig}, key.PublicKey)
}

func NewUnsignedTx(chainID *big.Int, nonce uint64, to Address, value Amount, gasLimit uint64, gasPrice Amount, data []byte) Tx {
	return &ethTx{
		chainID: chainID,
		tx:      types.NewTransaction(nonce, common.Address(to), value.ToBig(), gasLimit, gasPrice.ToBig(), data),
		signed:  false,
	}
}

func NewSignedTx(chainID *big.Int, tx *types.Transaction) Tx {
	return &ethTx{
		chainID: chainID,
		tx:      tx,
		signed:  true,
	}
}
