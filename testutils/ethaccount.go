package testutils

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/renproject/mercury/types/ethtypes"
)

func NewAccount() (*ecdsa.PrivateKey, ethtypes.Address, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, ethtypes.Address{}, err
	}
	addr := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	return privateKey, ethtypes.HexStringToAddress(addr), err
}

func NewAccountFromHexPrivateKey(hexString string) (*ecdsa.PrivateKey, ethtypes.Address, error) {
	if hasHexPrefix(hexString) {
		hexString = hexString[2:]
	}
	privateKey, err := crypto.HexToECDSA(hexString)
	if err != nil {
		return nil, ethtypes.Address{}, err
	}
	addr := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	return privateKey, ethtypes.HexStringToAddress(addr), err
}

func hasHexPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}
