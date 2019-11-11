package erc20

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/renproject/mercury/sdk/client/ethclient"
	"github.com/renproject/mercury/types/ethtypes"
)

// ERC20 is a bindings object for Erc20
type ERC20 interface {
	Balance(ctx context.Context, address ethtypes.Address) (*big.Int, error)
	TotalSupply(ctx context.Context) (*big.Int, error)
	Decimals(ctx context.Context) (uint8, error)

	Transfer(ctx context.Context, signer, to ethtypes.Address, amount *big.Int) (ethtypes.Tx, error)
	Approve(ctx context.Context, signer, to ethtypes.Address, amount *big.Int) (ethtypes.Tx, error)
	TransferFrom(ctx context.Context, signer, from, to ethtypes.Address, amount *big.Int) (ethtypes.Tx, error)
}

type erc20 struct {
	contract ethtypes.Contract
}

var ABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"}]"
var BIN = "0x00"

// New returns a new instance of ERC20
func New(client ethclient.Client, addr ethtypes.Address) (ERC20, error) {
	contract, err := client.Contract(addr, []byte(ABI))
	if err != nil {
		return nil, fmt.Errorf("failed to bind erc20 at address=%v: %v", addr, err)
	}
	return &erc20{
		contract: contract,
	}, nil
}

// Deploy a new contract, the contract returned by this function can only be used after signing and submitting the tx returned by this function.
func Deploy(ctx context.Context, client ethclient.Client, from ethtypes.Address, value *big.Int, params ...interface{}) (ethtypes.Tx, ERC20, error) {
	contractBin, err := hex.DecodeString(BIN)
	if err != nil {
		return ethtypes.Tx{}, nil, err
	}
	tx, contract, err := ethtypes.DeployContract(ctx, client.EthClient(), []byte(ABI), contractBin, from, value, params...)
	if err != nil {
		return ethtypes.Tx{}, nil, err
	}
	return tx, &erc20{contract}, nil
}

// Balance returns the balance of an address
func (e *erc20) Balance(ctx context.Context, address ethtypes.Address) (*big.Int, error) {
	balance := new(*big.Int)
	if err := e.contract.Call(ctx, ethtypes.Address{}, &balance, "balanceOf", address); err != nil {
		return nil, err
	}
	return *balance, nil
}

// TotalSupply returns the total supply of an ERC20
func (e *erc20) TotalSupply(ctx context.Context) (amount *big.Int, err error) {
	totalSupply := new(*big.Int)
	if err := e.contract.Call(ctx, ethtypes.Address{}, &totalSupply, "totalSupply"); err != nil {
		return nil, err
	}
	return *totalSupply, nil
}

// Decimals returns the total supply of an ERC20
func (e *erc20) Decimals(ctx context.Context) (decimals uint8, err error) {
	err = e.contract.Call(ctx, ethtypes.Address{}, &decimals, "decimals")
	return
}

// Transfer the given amount of ERC20 to the`to` address
func (e *erc20) Transfer(ctx context.Context, signer, to ethtypes.Address, amount *big.Int) (ethtypes.Tx, error) {
	return e.contract.BuildTx(ctx, signer, "transfer", nil, to, amount)
}

// Approve the given amount of ERC20 to the `to` address from the `from` address
func (e *erc20) Approve(ctx context.Context, signer, to ethtypes.Address, amount *big.Int) (ethtypes.Tx, error) {
	return e.contract.BuildTx(ctx, signer, "approve", nil, to, amount)
}

// TransferFrom the given amount of ERC20 to the `to` address from the `from` address
func (e *erc20) TransferFrom(ctx context.Context, signer, from, to ethtypes.Address, amount *big.Int) (ethtypes.Tx, error) {
	return e.contract.BuildTx(ctx, signer, "transferFrom", nil, from, to, amount)
}
