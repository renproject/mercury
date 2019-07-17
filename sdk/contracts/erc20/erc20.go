package erc20

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/renproject/mercury/bindings"
	"github.com/renproject/mercury/sdk/client/ethclient"
	"github.com/renproject/mercury/types/ethtypes"
)

// ERC20 is a bindings object for Erc20
type ERC20 interface {
	Balance(address ethtypes.Address) (ethtypes.Amount, error)
	TotalSupply() (ethtypes.Amount, error)
	Decimals() (uint8, error)
}

type erc20 struct {
	erc20 *bindings.Erc20
}

// New returns a new instance of ERC20
func New(client ethclient.Client, addr ethtypes.Address) (ERC20, error) {
	e, err := bindings.NewErc20(common.Address(addr), bind.ContractBackend(client.EthClient()))
	if err != nil {
		return &erc20{}, fmt.Errorf("failed to bind erc20 at address=%v: %v", addr, err)
	}
	return &erc20{
		erc20: e,
	}, nil
}

// Balance returns the balance of an address
func (e *erc20) Balance(address ethtypes.Address) (ethtypes.Amount, error) {
	amount, err := e.erc20.BalanceOf(&bind.CallOpts{}, common.Address(address))
	if err != nil {
		return ethtypes.Amount{}, err
	}
	return ethtypes.WeiFromBig(amount), nil
}

// TotalSupply returns the total supply of an ERC20
func (e *erc20) TotalSupply() (ethtypes.Amount, error) {
	amount, err := e.erc20.TotalSupply(&bind.CallOpts{})
	if err != nil {
		return ethtypes.Amount{}, err
	}
	return ethtypes.WeiFromBig(amount), nil
}

// Decimals returns the total supply of an ERC20
func (e *erc20) Decimals() (uint8, error) {
	return e.erc20.Decimals(&bind.CallOpts{})
}
