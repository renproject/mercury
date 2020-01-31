package ethtypes

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"reflect"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type contract struct {
	abi     abi.ABI
	address Address
	client  *ethclient.Client
}

type Contract interface {
	Address() Address
	BuildTx(ctx context.Context, from Address, method string, value *big.Int, params ...interface{}) (Tx, error)
	Call(ctx context.Context, caller Address, result interface{}, method string, params ...interface{}) error
	Watch(ctx context.Context, events chan<- Event, beginBlockNum *big.Int, event string,
		indexedArgs ...Hash) error
}

type ContractRevert struct {
	ErrorMsg string
}

func (revert ContractRevert) Error() string {
	return revert.ErrorMsg
}

func NewContractRevert(err string) error {
	return ContractRevert{ErrorMsg: err}
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

func DeployContract(ctx context.Context, client *ethclient.Client, contractABI []byte, bin []byte, from Address, value *big.Int, params ...interface{}) (Tx, Contract, error) {
	abi, err := abi.JSON(bytes.NewBuffer(contractABI))
	if err != nil {
		return Tx{}, nil, err
	}
	c := &contract{
		client:  client,
		address: Address{},
		abi:     abi,
	}
	tx, err := c.buildTx(ctx, from, bin, "", value, params...)
	return tx, c, err
}

func (c *contract) BuildTx(ctx context.Context, from Address, method string, value *big.Int, params ...interface{}) (Tx, error) {
	return c.buildTx(ctx, from, nil, method, value, params...)
}

func (c *contract) Address() Address {
	return c.address
}

func (c *contract) buildTx(ctx context.Context, from Address, bin []byte, method string, value *big.Int, params ...interface{}) (Tx, error) {
	data, err := c.abi.Pack(method, params...)
	if err != nil {
		return Tx{}, fmt.Errorf("failed to pack data: %v", err)
	}

	if (c.address == Address{}) {
		if bin == nil {
			return Tx{}, fmt.Errorf("failed to deploy a contract: contract bin not provided")
		}
		data = append(bin, data...)
	}

	// Ensure a valid value field and resolve the account nonce
	if value == nil {
		value = new(big.Int)
	}

	nonce, err := c.client.NonceAt(ctx, common.Address(from), nil)
	if err != nil {
		return Tx{}, fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	// Figure out the gas allowance and gas price values
	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		return Tx{}, fmt.Errorf("failed to suggest gas price: %v", err)
	}

	chainID, err := c.client.ChainID(ctx)
	if err != nil {
		return Tx{}, fmt.Errorf("failed to get chain id: %v", err)
	}

	// Create the transaction, sign it and schedule it for execution
	var rawTx *types.Transaction
	if (c.address == Address{}) {
		// If the contract surely has code (or code is not needed), estimate the transaction
		msg := ethereum.CallMsg{From: common.Address(from), To: nil, Value: value, Data: data}
		c.address = Address(crypto.CreateAddress(common.Address(from), nonce))

		gasLimit, err := c.client.EstimateGas(ctx, msg)
		if err != nil {
			rawTx = types.NewContractCreation(nonce, value, 2500000, gasPrice, data)
			return Tx{chainID, rawTx, false}, NewContractRevert(fmt.Sprintf("failed to estimate gas needed: %v", err))
		}
		rawTx = types.NewContractCreation(nonce, value, gasLimit, gasPrice, data)
	} else {
		contractAddr := common.Address(c.address)
		// If the contract surely has code (or code is not needed), estimate the transaction
		msg := ethereum.CallMsg{From: common.Address(from), To: &contractAddr, Gas: 7000000, Value: value, Data: data}

		output, err := c.client.CallContract(ctx, msg, nil)
		if err != nil {
			return Tx{}, fmt.Errorf("failed to simulate the transaction: %v", err)
		}

		gasLimit, err := c.client.EstimateGas(ctx, msg)
		if err != nil {
			if len(output) != 0 {
				if reason, err := parseRevertReason(output); err == nil {
					rawTx = types.NewTransaction(nonce, contractAddr, value, 2500000, gasPrice, data)
					return Tx{chainID, rawTx, false}, NewContractRevert(reason)
				}
			}
			rawTx = types.NewTransaction(nonce, contractAddr, value, 2500000, gasPrice, data)
			return Tx{chainID, rawTx, false}, NewContractRevert(fmt.Sprintf("failed to estimate gas needed: %v", err))
		}
		rawTx = types.NewTransaction(nonce, contractAddr, value, gasLimit, gasPrice, data)
	}

	return Tx{
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
	msg := ethereum.CallMsg{To: &contractAddr, Data: input}
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
		return fmt.Errorf("error calling %s: %v", method, err)
	}

	switch result := result.(type) {
	case *Address:
		var resAddr common.Address
		if err := c.abi.Unpack(&resAddr, method, output); err != nil {
			return err
		}
		return set(result, AddressFromHex(resAddr.Hex()))
	default:
		return c.abi.Unpack(result, method, output)
	}
}

// set attempts to assign src to dst by either setting, copying or otherwise.
//
// set is a bit more lenient when it comes to assignment and doesn't force an as
// strict ruleset as bare `reflect` does.
func set(dstVal, srcVal interface{}) error {
	dst := reflect.ValueOf(dstVal).Elem()
	src := reflect.ValueOf(srcVal)

	dstType, srcType := dst.Type(), src.Type()
	switch {
	case dstType.Kind() == reflect.Interface && dst.Elem().IsValid():
		return set(dst.Elem(), src)
	case srcType.AssignableTo(dstType) && dst.CanSet():
		dst.Set(src)
	default:
		return fmt.Errorf("abi: cannot unmarshal %v in to %v", src.Type(), dst.Type())
	}
	return nil
}

func (c *contract) Watch(ctx context.Context, events chan<- Event, beginBlockNum *big.Int, event string, indexedArgs ...Hash) error {
	if beginBlockNum == nil {
		beginBlockNum = big.NewInt(0)
	}
	ticker := time.NewTicker(10 * time.Second)
	for {
		topics, err := c.getTopics(event, indexedArgs)
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			logs, err := c.client.FilterLogs(ctx, ethereum.FilterQuery{
				FromBlock: beginBlockNum,
				Addresses: []common.Address{
					common.Address(c.address),
				},
				Topics: [][]common.Hash{topics},
			})
			if err != nil {
				return fmt.Errorf("failed to filter logs: %v", err)
			}
			for _, log := range logs {
				beginBlockNum.SetUint64(log.BlockNumber + 1)
				event, err := c.abi.EventByID(log.Topics[0])
				if err != nil {
					return fmt.Errorf("failed to get event by id: %s", err)
				}
				eventArgs := map[string]interface{}{}
				if err := event.Inputs.UnpackIntoMap(eventArgs, log.Data); err != nil {
					return fmt.Errorf("failed to unpack an event: %v", err)
				}
				// Try to use the timestamp of the log, use zero on failure.
				header, err := c.client.HeaderByHash(ctx, log.BlockHash)
				if err != nil {
					return fmt.Errorf("failed to get event time: %v", err)
				}
				events <- Event{
					Name:        event.Name,
					TxHash:      TxHash(log.TxHash),
					IndexedArgs: decodeHashes(log.Topics[1:]),
					Args:        eventArgs,

					Timestamp:   header.Time,
					BlockNumber: log.BlockNumber,
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

func parseRevertReason(data []byte) (string, error) {
	args := abi.Arguments{{Type: abi.Type{
		Kind: reflect.String,
		Type: reflect.TypeOf(""),
		T:    abi.StringTy,
	}}}
	revertReason := ""
	if err := args.Unpack(&revertReason, data[4:]); err != nil {
		return "", err
	}
	return revertReason, nil
}

func (c *contract) getTopics(event string, indexedArgs []Hash) ([]common.Hash, error) {
	eventABI := c.abi.Events[event]
	topics := []common.Hash{eventABI.ID()}
	if len(indexedArgs) > len(eventABI.Inputs)-eventABI.Inputs.LengthNonIndexed() {
		return topics, fmt.Errorf("invalid number of indexed args: %v", len(indexedArgs))
	}
	for _, arg := range indexedArgs {
		topics = append(topics, common.Hash(arg))
	}
	return topics, nil
}
