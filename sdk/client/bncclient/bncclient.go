package bncclient

import (
	"encoding/hex"
	"fmt"

	"github.com/binance-chain/go-sdk/client/basic"
	"github.com/binance-chain/go-sdk/client/query"
	"github.com/binance-chain/go-sdk/client/websocket"
	"github.com/binance-chain/go-sdk/types/msg"
	"github.com/binance-chain/go-sdk/types/tx"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/bnctypes"
)

type Client interface {
	PrintTime()
	Balances(from bnctypes.Address) (bnctypes.Coins, error)
	Mint(from bnctypes.Address, coin bnctypes.Coin) (types.Tx, error)
	Burn(from bnctypes.Address, coin bnctypes.Coin) (types.Tx, error)
	Send(from bnctypes.Address, recipients bnctypes.Recipients) (types.Tx, error)
	SubmitTx(tx types.Tx) error
}

type client struct {
	network     bnctypes.Network
	queryClient query.QueryClient
	basicClient basic.BasicClient
	wsClient    websocket.WSClient
}

func New(network bnctypes.Network) Client {
	var baseURL string
	switch network {
	case bnctypes.Mainnet:
		baseURL = "dex.binance.org"
	case bnctypes.Testnet:
		baseURL = "testnet-dex.binance.org"
	default:
		panic(types.ErrUnknownNetwork)
	}

	c := basic.NewClient(baseURL)
	return &client{
		network:     network,
		basicClient: c,
		queryClient: query.NewClient(c),
		wsClient:    websocket.NewClient(c),
	}
}

func (client *client) PrintTime() {
	t, err := client.queryClient.GetTime()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(t.ApTime, t.BlockTime)
}

func (client *client) Mint(from bnctypes.Address, coin bnctypes.Coin) (types.Tx, error) {
	return client.BuildTx(from, msg.NewMintMsg(
		from.AccAddress(),
		coin.Denom,
		coin.Amount,
	))
}

func (client *client) Burn(from bnctypes.Address, coin bnctypes.Coin) (types.Tx, error) {
	return client.BuildTx(from, msg.NewTokenBurnMsg(
		from.AccAddress(),
		coin.Denom,
		coin.Amount,
	))
}

func (client *client) Balances(from bnctypes.Address) (bnctypes.Coins, error) {
	acc, err := client.queryClient.GetAccount(from.String())
	if err != nil {
		return nil, err
	}
	balances := make([]bnctypes.Coin, len(acc.Balances))
	for i := range balances {
		balances[i] = bnctypes.Coin{
			Denom:  acc.Balances[i].Symbol,
			Amount: acc.Balances[i].Free.ToInt64(),
		}
	}
	return bnctypes.NewCoins(balances...), nil
}

func (client *client) Send(from bnctypes.Address, recipients bnctypes.Recipients) (types.Tx, error) {
	return client.BuildTx(from, recipients.SendTx(from))
}

func (client *client) BuildTx(from bnctypes.Address, m msg.Msg) (types.Tx, error) {
	acc, err := client.queryClient.GetAccount(from.String())
	if err != nil {
		return nil, err
	}

	// prepare message to sign
	signMsg := tx.StdSignMsg{
		ChainID:       client.network.ChainID(),
		AccountNumber: acc.Number,
		Sequence:      acc.Sequence,
		Memo:          "",
		Msgs:          []msg.Msg{m},
		Source:        tx.Source,
	}

	// special logic for createOrder, to save account query
	if orderMsg, ok := m.(msg.CreateOrderMsg); ok {
		orderMsg.ID = msg.GenerateOrderID(signMsg.Sequence+1, from.AccAddress())
		signMsg.Msgs[0] = orderMsg
	}

	// validate messages
	for _, m := range signMsg.Msgs {
		if err := m.ValidateBasic(); err != nil {
			return nil, err
		}
	}

	return bnctypes.NewTx(signMsg), nil
}

func (client *client) SubmitTx(tx types.Tx) error {
	stx, err := tx.Serialize()
	if err != nil {
		return err
	}
	params := map[string]string{}
	params["sync"] = "true"
	if _, err := client.basicClient.PostTx([]byte(hex.EncodeToString(stx)), params); err != nil {
		return err
	}
	return nil
}
