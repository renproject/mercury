package btcaccount

import (
	"context"
	"crypto/ecdsa"

	"github.com/btcsuite/btcutil"
	"github.com/renproject/mercury/sdk/client/btcclient"
	"github.com/renproject/mercury/types/btctypes"
	"github.com/sirupsen/logrus"
)

type Transaction struct {
	TxId               string `json:"txid"`
	SourceAddress      string `json:"source_address"`
	DestinationAddress string `json:"destination_address"`
	Amount             int64  `json:"amount"`
	UnsignedTx         string `json:"unsignedtx"`
	SignedTx           string `json:"signedtx"`
}

type Account struct {
	Client *btcclient.Client

	logger logrus.FieldLogger
	key *ecdsa.PrivateKey
}

func NewBtcAccount (logger logrus.FieldLogger, client *btcclient.Client, key *ecdsa.PrivateKey) *Account{
	return &Account{
		Client: client,
		logger: logger,
		key:    key,
	}
}

func NewAccountFromWIF(logger logrus.FieldLogger, client *btcclient.Client, wifStr string) (*Account, error){
	wif, err:= btcutil.DecodeWIF(wifStr )
	if err != nil {
		return nil, err
	}
	privKey := (*ecdsa.PrivateKey)(wif.PrivKey)
	return &Account{
		Client: client,
		logger: logger,
		key:    privKey,
	}, nil
}

func (account Account) Address() (btctypes.Addr, error){
	return btctypes.AddressFromPubKey(&account.key.PublicKey, account.Client.Network)
}

func (account Account) Transfer(ctx context.Context, to btctypes.Addr, value, fees btctypes.Amount) error {
	panic("unimplemented")
	// senderAddr, err := account.Address()
	// if err != nil {
	// 	return err
	// }
	//
	// // Get utxos
	// utxos, err := account.Client.UTXOs(ctx, senderAddr, 999999, 0 )
	// if err != nil {
	// 	return err
	// }
	//
	// // Construct the tx.
	// sourceTx := wire.NewMsgTx(wire.TxVersion)
	// sourceUtxoHash, err := chainhash.NewHashFromStr(utxos[0].TxHash)
	// if err != nil {
	// 	return err
	// }
	// sourceUtxo := wire.NewOutPoint(sourceUtxoHash, 0)
	// sourceTxIn := wire.NewTxIn(sourceUtxo, nil, nil)
	//
	//
	// destinationPkScript, err := txscript.PayToAddrScript(to)
	// if err != nil {
	// 	return err
	// }
	// sourcePkScript, err := txscript.PayToAddrScript(senderAddr)
	// if err != nil {
	// 	return err
	// }
	//
	// sourceTxOut := wire.NewTxOut(int64(value), sourcePkScript)
	// sourceTx.AddTxIn(sourceTxIn)
	// sourceTx.AddTxOut(sourceTxOut)
	// sourceTxHash := sourceTx.TxHash()
	// redeemTx := wire.NewMsgTx(wire.TxVersion)
	// prevOut := wire.NewOutPoint(&sourceTxHash, 0)
	// redeemTxIn := wire.NewTxIn(prevOut, nil, nil)
	// redeemTx.AddTxIn(redeemTxIn)
	// redeemTxOut := wire.NewTxOut(int64(value), destinationPkScript)
	// redeemTx.AddTxOut(redeemTxOut)
	// sigScript, err := txscript.SignatureScript(redeemTx, 0, sourceTx.TxOut[0].PkScript, txscript.SigHashAll, wif.PrivKey, false)
	// if err != nil {
	// 	return err
	// }
	// redeemTx.TxIn[0].SignatureScript = sigScript
	// flags := txscript.StandardVerifyFlags
	// vm, err := txscript.NewEngine(sourceTx.TxOut[0].PkScript, redeemTx, 0, flags, nil, nil, amount)
	// if err != nil {
	// 	return  err
	// }
	// if err := vm.Execute(); err != nil {
	// 	return  err
	// }
	// var unsignedTx bytes.Buffer
	// var signedTx bytes.Buffer
	// sourceTx.Serialize(&unsignedTx)
	// redeemTx.Serialize(&signedTx)
	// transaction.TxId = sourceTxHash.String()
	// transaction.UnsignedTx = hex.EncodeToString(unsignedTx.Bytes())
	// transaction.Amount = amount
	// transaction.SignedTx = hex.EncodeToString(signedTx.Bytes())
	// transaction.SourceAddress = sourceAddress.EncodeAddress()
	// transaction.DestinationAddress = destinationAddress.EncodeAddress()
	// return transaction, nil
}


