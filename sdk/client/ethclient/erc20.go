package ethclient

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/renproject/mercury/bindings"
	"github.com/renproject/mercury/types/ethtypes"
)

// ERC20 is a bindings object for Erc20
type ERC20 struct {
	erc20 *bindings.Erc20
}

// Balance returns the balance of an address
func (e *ERC20) Balance(address ethtypes.Address) (ethtypes.Amount, error) {
	amount, err := e.erc20.BalanceOf(&bind.CallOpts{}, common.Address(address))
	if err != nil {
		return ethtypes.Amount{}, err
	}
	return ethtypes.WeiFromBig(amount), nil
}

// TotalSupply returns the total supply of an ERC20
func (e *ERC20) TotalSupply() (ethtypes.Amount, error) {
	amount, err := e.erc20.TotalSupply(&bind.CallOpts{})
	if err != nil {
		return ethtypes.Amount{}, err
	}
	return ethtypes.WeiFromBig(amount), nil
}

// Decimals returns the total supply of an ERC20
func (e *ERC20) Decimals() (uint8, error) {
	return e.erc20.Decimals(&bind.CallOpts{})
}

func erc20Token(client *ethclient.Client, addr ethtypes.Address) (ERC20, error) {
	e, err := bindings.NewErc20(common.Address(addr), bind.ContractBackend(client))
	if err != nil {
		return ERC20{}, fmt.Errorf("failed to bind erc20 at address=%v: %v", addr, err)
	}
	return ERC20{
		erc20: e,
	}, nil
}
