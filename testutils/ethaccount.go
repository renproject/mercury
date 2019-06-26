package testutils

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/renproject/mercury/types/ethtypes"
)

func NewAccount() (*ecdsa.PrivateKey, ethtypes.EthAddr, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, ethtypes.EthAddr{}, err
	}
	addr := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	return privateKey, ethtypes.HexStringToEthAddr(addr), err
}
