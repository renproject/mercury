package ethtypes

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	coretypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/renproject/mercury/types"
)

const (
	Mainnet     network = 1
	Ropsten     network = 3
	Kovan       network = 42
	Ganache     network = 255
	EthLocalnet network = 254
)

func (network network) String() string {
	switch network {
	case Mainnet:
		return "mainnet"
	case Kovan:
		return "kovan"
	case Ganache:
		return "ganache"
	case EthLocalnet:
		return "localnet"
	default:
		panic(types.ErrUnknownNetwork)
	}
}

func (network network) Chain() types.Chain {
	return types.Ethereum
}

type Network interface {
	types.Network
}

type network uint8

type Tx struct {
	chainID *big.Int
	tx      *coretypes.Transaction
	signed  bool
}

type TxHash common.Hash

func (txHash TxHash) String() string {
	return common.Hash(txHash).String()
}

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
	if tx.tx == nil {
		panic("pre-condition violation: cannot sign a nil transaction")
	}

	signer := coretypes.NewEIP155Signer(tx.chainID)
	signedTx, err := coretypes.SignTx(tx.tx, signer, key)
	if err != nil {
		return err
	}

	tx.tx = signedTx
	tx.signed = true
	return nil
}

func (tx *Tx) SetNonce(newNonce uint64) {
	if tx.tx.To() == nil {
		tx.tx = coretypes.NewContractCreation(newNonce, tx.tx.Value(), tx.tx.Gas(), tx.tx.GasPrice(), tx.tx.Data())
	} else {
		tx.tx = coretypes.NewTransaction(newNonce, *tx.tx.To(), tx.tx.Value(), tx.tx.Gas(), tx.tx.GasPrice(), tx.tx.Data())
	}
	tx.signed = false
}

func (tx *Tx) SetGasPrice(newGasPrice *big.Int) {
	if tx.tx.To() == nil {
		tx.tx = coretypes.NewContractCreation(tx.tx.Nonce(), tx.tx.Value(), tx.tx.Gas(), newGasPrice, tx.tx.Data())
	} else {
		tx.tx = coretypes.NewTransaction(tx.tx.Nonce(), *tx.tx.To(), tx.tx.Value(), tx.tx.Gas(), newGasPrice, tx.tx.Data())
	}
	tx.signed = false
}

func (tx *Tx) SetGas(newGas uint64) {
	if tx.tx.To() == nil {
		tx.tx = coretypes.NewContractCreation(tx.tx.Nonce(), tx.tx.Value(), newGas, tx.tx.GasPrice(), tx.tx.Data())
	} else {
		tx.tx = coretypes.NewTransaction(tx.tx.Nonce(), *tx.tx.To(), tx.tx.Value(), newGas, tx.tx.GasPrice(), tx.tx.Data())
	}
	tx.signed = false
}

func NewUnsignedTx(chainID *big.Int, nonce uint64, to *Address, value Amount, gasLimit uint64, gasPrice Amount, data []byte) Tx {
	tx := new(coretypes.Transaction)
	if to == nil {
		tx = coretypes.NewContractCreation(nonce, value.ToBig(), gasLimit, gasPrice.ToBig(), data)
	} else {
		tx = coretypes.NewTransaction(nonce, common.Address(*to), value.ToBig(), gasLimit, gasPrice.ToBig(), data)
	}
	return Tx{
		chainID: chainID,
		tx:      tx,
		signed:  false,
	}
}

func NewSignedTx(tx *coretypes.Transaction) Tx {
	return Tx{
		chainID: tx.ChainId(),
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

type Hash common.Hash

func HashFromHex(hashStr string) Hash {
	return Hash(common.HexToHash(hashStr))
}

func HashFromBytes(hashBytes []byte) Hash {
	return Hash(common.BytesToHash(hashBytes))
}

type Event struct {
	Name        string
	Args        map[string]interface{}
	IndexedArgs []Hash

	Timestamp   uint64
	TxHash      TxHash
	BlockNumber uint64
}

func Keccak256(data interface{}) Hash {
	switch data := data.(type) {
	case string:
		return HashFromBytes(crypto.Keccak256([]byte(data)))
	case []byte:
		return HashFromBytes(crypto.Keccak256(data))
	case [][]byte:
		return HashFromBytes(crypto.Keccak256(data...))
	default:
		panic(fmt.Sprintf("unsupported type: %T", data))
	}
}
