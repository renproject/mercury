package testutils

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
)

func NewAccount() (privateKey *ecdsa.PrivateKey, publicAddr string, err error) {
	privateKey, err = crypto.GenerateKey()
	publicAddr = crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	return privateKey, publicAddr, err
}
