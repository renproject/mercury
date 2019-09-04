package types

import (
	"crypto/ecdsa"
)

type SignatureHash []byte

type TxHash string

type Tx interface {
	SignatureHashes() []SignatureHash
	Sign(key *ecdsa.PrivateKey) (err error)
	IsSigned() bool
	Serialize() ([]byte, error)
	Hash() TxHash
	InjectSigs(sigs [][]byte, pubKey ecdsa.PublicKey) error
}
