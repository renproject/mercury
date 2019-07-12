package types

import (
	"crypto/ecdsa"

	"github.com/btcsuite/btcd/btcec"
)

type SignatureHash []byte

type TxHash string

type Tx interface {
	SignatureHashes() []SignatureHash
	Sign(key *ecdsa.PrivateKey) (err error)
	IsSigned() bool
	Serialize() ([]byte, error)
	Hash() TxHash
	InjectSignatures(sigs []*btcec.Signature, serializedPubKey []byte) error
}

type Confirmations int64
