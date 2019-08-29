package bncclient

import (
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
	Balances(from bnctypes.Address) (int64, error)
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
	case bnctypes.Testnet:
		baseURL = "dex.binance.org"
	case bnctypes.Mainnet:
		baseURL = "testnet-dex.binance.org"
	default:
		panic(types.ErrUnknownNetwork)
	}

	c := basic.NewClient(baseURL)
	return &client{
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

func (client *client) OpenOrder(from bnctypes.Address) {
	return client.BuildTx(from, msg.NewCreateOrderMsg(from.AccAddress(), "", 0))
}

func (client *client) Mint(from bnctypes.Address, symbol string, amount int64) (types.Tx, error) {
	return client.BuildTx(from, msg.NewMintMsg(
		from.AccAddress(),
		symbol,
		amount,
	))
}

func (client *client) BuildTx(from bnctypes.Address, m msg.Msg) (types.Tx, error) {
	acc, err := client.queryClient.GetAccount(from.String())
	if err != nil {
		return nil, err
	}

	// prepare message to sign
	signMsg := tx.StdSignMsg{
		ChainID:       client.network.String(),
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
