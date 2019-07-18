package ethtypes

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	coretypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/renproject/mercury/types"
)

const (
	Mainnet Network = 1
	Kovan   Network = 42
)

func (network Network) String() string {
	switch network {
	case Mainnet:
		return "mainnet"
	case Kovan:
		return "kovan"
	default:
		panic(types.ErrUnknownNetwork)
	}
}

type Network uint8

type Tx struct {
	chainID *big.Int
	tx      *coretypes.Transaction
	signed  bool
}

type TxHash common.Hash

func NewTxHashFromHex(hexString string) TxHash {
	return TxHash(common.HexToHash(hexString))
}

func (tx *Tx) Hash() TxHash {
	return TxHash(tx.tx.Hash())
}

func (tx *Tx) IsSigned() bool {
	return tx.signed
}

func (tx *Tx) ToTransaction() *coretypes.Transaction {
	return tx.tx
}

func (tx *Tx) Sign(key *ecdsa.PrivateKey) error {
	// Pre-condition checks
	if tx.IsSigned() {
		panic("pre-condition violation: cannot sign already signed transaction")
	}

	signer := coretypes.NewEIP155Signer(tx.chainID)
	signedTx, err := coretypes.SignTx((*coretypes.Transaction)(tx.tx), signer, key)
	if err != nil {
		return err
	}

	tx.tx = signedTx
	tx.signed = true
	return nil
}

func NewUnsignedTx(chainID *big.Int, nonce uint64, to Address, value Amount, gasLimit uint64, gasPrice Amount, data []byte) Tx {
	return Tx{
		chainID: chainID,
		tx:      coretypes.NewTransaction(nonce, common.Address(to), value.ToBig(), gasLimit, gasPrice.ToBig(), data),
		signed:  false,
	}
}

func NewSignedTx(chainID *big.Int, tx *coretypes.Transaction) Tx {
	return Tx{
		chainID: chainID,
		tx:      tx,
		signed:  true,
	}
}

type Address common.Address

func AddressFromPublicKey(publicKeyECDSA *ecdsa.PublicKey) Address {
	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return Address(address)
}

func AddressFromHex(addr string) Address {
	return Address(common.HexToAddress(addr))
}

func (addr Address) Hex() string {
	return common.Address(addr).Hex()
}
