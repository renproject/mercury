package zec

import (
	"fmt"
)

type multiClient struct {
	initiated bool
	clients   []ZCashClient
}

func NewMulti(clients ...ZCashClient) ZCashClient {
	return &multiClient{
		clients: clients,
	}
}

func (zec *multiClient) Init() error {
	for _, client := range zec.clients {
		if err := client.Init(); err != nil {
			return err
		}
	}
	zec.initiated = true
	return nil
}

func (zec *multiClient) GetUTXOs(address string, limit, confitmations int64) ([]UTXO, error) {
	for i, client := range zec.clients {
		if utxos, err := client.GetUTXOs(address, limit, confitmations); i+1 == len(zec.clients) {
			return utxos, err
		}
	}
	return []UTXO{}, fmt.Errorf("no clients provided")
}

func (zec *multiClient) GetUTXO(txHash string, vout int64) (UTXO, error) {
	return UTXO{}, fmt.Errorf("no clients provided")
}

func (zec *multiClient) Confirmations(txHashStr string) (int64, error) {
	for i, client := range zec.clients {
		if conf, err := client.Confirmations(txHashStr); err == nil || i+1 == len(zec.clients) {
			return conf, err
		}
	}
	return 0, fmt.Errorf("no clients provided")
}

func (zec *multiClient) ScriptFunded(address string, value int64) (bool, int64, error) {
	for i, client := range zec.clients {
		if funded, val, err := client.ScriptFunded(address, value); err == nil || i+1 == len(zec.clients) {
			return funded, val, err
		}
	}
	return false, 0, fmt.Errorf("no clients provided")
}

func (zec *multiClient) ScriptRedeemed(address string, value int64) (bool, int64, error) {
	for i, client := range zec.clients {
		if redeemed, val, err := client.ScriptRedeemed(address, value); err == nil || i+1 == len(zec.clients) {
			return redeemed, val, err
		}
	}
	return false, 0, fmt.Errorf("no clients provided")
}

func (zec *multiClient) ScriptSpent(scriptAddress, spenderAddress string) (bool, string, error) {
	for i, client := range zec.clients {
		if spent, val, err := client.ScriptSpent(scriptAddress, spenderAddress); err == nil || i+1 == len(zec.clients) {
			return spent, val, err
		}
	}
	return false, "", fmt.Errorf("no clients provided")
}

func (zec *multiClient) PublishTransaction(stx []byte) error {
	for i, client := range zec.clients {
		if err := client.PublishTransaction(stx); err == nil || i+1 == len(zec.clients) {
			return err
		}
	}
	return fmt.Errorf("no clients provided")
}

func (zec *multiClient) Health() bool {
	var health bool
	for _, client := range zec.clients {
		health = health || client.Health()
	}
	return health
}
