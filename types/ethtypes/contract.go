package ethtypes

import (
	"bytes"
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type contract struct {
	abi     abi.ABI
	address Address
	client  *ethclient.Client
}

type Contract interface {
	BuildTx(ctx context.Context, from Address, method string, value *big.Int, params ...interface{}) (Tx, error)
	Call(ctx context.Context, caller Address, result interface{}, method string, params ...interface{}) error
	Watch(ctx context.Context, events chan<- Event, beginBlockNum *big.Int, topics ...[]Hash) error
}

func NewContract(client *ethclient.Client, address Address, contractABI []byte) (Contract, error) {
	abi, err := abi.JSON(bytes.NewBuffer(contractABI))
	if err != nil {
		return nil, err
	}
	return &contract{
		client:  client,
		address: address,
		abi:     abi,
	}, nil
}

func (c *contract) BuildTx(ctx context.Context, from Address, method string, value *big.Int, params ...interface{}) (Tx, error) {

	data, err := c.abi.Pack(method, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to pack data: %v", err)
	}

	// Ensure a valid value field and resolve the account nonce
	if value == nil {
		value = new(big.Int)
	}

	nonce, err := c.client.PendingNonceAt(ctx, common.Address(from))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	// Figure out the gas allowance and gas price values
	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to suggest gas price: %v", err)
	}

	contractAddr := common.Address(c.address)
	// If the contract surely has code (or code is not needed), estimate the transaction
	msg := ethereum.CallMsg{From: common.Address(from), To: &contractAddr, Value: value, Data: data}
	gasLimit, err := c.client.EstimateGas(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas needed: %v", err)
	}

	// Create the transaction, sign it and schedule it for execution
	var rawTx *types.Transaction
	if (c.address == Address{}) {
		rawTx = types.NewContractCreation(nonce, value, gasLimit, gasPrice, data)
	} else {
		rawTx = types.NewTransaction(nonce, contractAddr, value, gasLimit, gasPrice, data)
	}

	chainID, err := c.client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain id: %v", err)
	}

	return &ethTx{
		tx:      rawTx,
		chainID: chainID,
	}, nil
}

func (c *contract) Call(ctx context.Context, caller Address, result interface{}, method string, params ...interface{}) error {
	// Pack the input, call and unpack the results
	input, err := c.abi.Pack(method, params...)
	if err != nil {
		return err
	}
	contractAddr := common.Address(c.address)
	msg := ethereum.CallMsg{From: common.Address(caller), To: &contractAddr, Data: input}
	output, err := c.client.CallContract(ctx, msg, nil)
	if err == nil && len(output) == 0 {
		// Make sure we have a contract to operate on, and bail out otherwise.
		if code, err := c.client.CodeAt(ctx, contractAddr, nil); err != nil {
			return err
		} else if len(code) == 0 {
			return fmt.Errorf("no code")
		}
	}
	if err != nil {
		return err
	}

	return c.abi.Unpack(result, method, output)
}

func (c *contract) Watch(ctx context.Context, events chan<- Event, beginBlockNum *big.Int, topics ...[]Hash) error {
	if beginBlockNum == nil {
		beginBlockNum = big.NewInt(0)
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			logs, err := c.client.FilterLogs(ctx, ethereum.FilterQuery{
				FromBlock: beginBlockNum,
				Addresses: []common.Address{
					common.Address(c.address),
				},
				Topics: encodeHashMatrix(topics),
			})
			if err != nil {
				return fmt.Errorf("failed to filter logs")
			}
			for _, log := range logs {
				beginBlockNum.SetUint64(log.BlockNumber)
				event, err := c.abi.EventByID(log.Topics[0])
				if err != nil {
					return fmt.Errorf("failed to get event by id: %s", err)
				}
				eventArgs := map[string]interface{}{}
				if err := event.Inputs.UnpackIntoMap(eventArgs, log.Data); err != nil {
					return fmt.Errorf("failed to unpack an event: %v", err)
				}
				events <- Event{
					Name:        event.Name,
					IndexedArgs: decodeHashes(log.Topics[1:]),
					Args:        eventArgs,
				}
			}
		}
	}
}

func encodeHashMatrix(hashMatrix [][]Hash) [][]common.Hash {
	commonHashes := make([][]common.Hash, len(hashMatrix))
	for i, hashRow := range hashMatrix {
		commonHashes[i] = make([]common.Hash, len(hashRow))
		for j, hashColumn := range hashRow {
			commonHashes[i][j] = common.Hash(hashColumn)
		}
	}
	return commonHashes
}

func decodeHashes(chashes []common.Hash) []Hash {
	hashes := make([]Hash, len(chashes))
	for i, chash := range chashes {
		hashes[i] = Hash(chash)
	}
	return hashes
}
