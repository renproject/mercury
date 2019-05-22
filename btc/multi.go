package btc

import (
	"context"
	"fmt"
)

type multiClient struct {
	initiated bool
	clients   []BitcoinClient
}

func NewMulti(clients ...BitcoinClient) BitcoinClient {
	return &multiClient{
		clients: clients,
	}
}

func (btc *multiClient) Init() error {
	for _, client := range btc.clients {
		if err := client.Init(); err != nil {
			return err
		}
	}
	btc.initiated = true
	return nil
}

func (btc *multiClient) GetUTXOs(ctx context.Context, address string, limit, confitmations int) ([]UTXO, error) {
	for i, client := range btc.clients {
		if utxos, err := client.GetUTXOs(ctx, address, limit, confitmations); (err == nil && len(utxos) > 0) || i+1 == len(btc.clients) {
			return utxos, err
		}
	}
	return []UTXO{}, fmt.Errorf("no clients provided")
}

func (btc *multiClient) Confirmations(ctx context.Context, txHashStr string) (int64, error) {
	for i, client := range btc.clients {
		if conf, err := client.Confirmations(ctx, txHashStr); err == nil || i+1 == len(btc.clients) {
			return conf, err
		}
	}
	return 0, fmt.Errorf("no clients provided")
}

func (btc *multiClient) ScriptFunded(ctx context.Context, address string, value int64) (bool, int64, error) {
	for i, client := range btc.clients {
		if funded, val, err := client.ScriptFunded(ctx, address, value); err == nil || i+1 == len(btc.clients) {
			return funded, val, err
		}
	}
	return false, 0, fmt.Errorf("no clients provided")
}

func (btc *multiClient) ScriptRedeemed(ctx context.Context, address string, value int64) (bool, int64, error) {
	for i, client := range btc.clients {
		if redeemed, val, err := client.ScriptRedeemed(ctx, address, value); err == nil || i+1 == len(btc.clients) {
			return redeemed, val, err
		}
	}
	return false, 0, fmt.Errorf("no clients provided")
}

func (btc *multiClient) ScriptSpent(ctx context.Context, scriptAddress, spenderAddress string) (bool, string, error) {
	for i, client := range btc.clients {
		if spent, val, err := client.ScriptSpent(ctx, scriptAddress, spenderAddress); err == nil || i+1 == len(btc.clients) {
			return spent, val, err
		}
	}
	return false, "", fmt.Errorf("no clients provided")
}

func (btc *multiClient) PublishTransaction(ctx context.Context, stx []byte) error {
	for i, client := range btc.clients {
		if err := client.PublishTransaction(ctx, stx); err == nil || i+1 == len(btc.clients) {
			return err
		}
	}
	return fmt.Errorf("no clients provided")
}

func (btc *multiClient) OmniGetBalance(token int64, address string) (OmniGetBalanceResponse, error) {
	for i, client := range btc.clients {
		if bal, err := client.OmniGetBalance(token, address); err == nil || i+1 == len(btc.clients) {
			return bal, err
		}
	}
	return OmniGetBalanceResponse{}, fmt.Errorf("no clients provided")
}

func (btc *multiClient) Health() bool {
	var health bool
	for _, client := range btc.clients {
		health = health || client.Health()
	}
	return health
}
