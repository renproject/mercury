package proxy

import (
	"github.com/renproject/mercury/rpc/btcrpc"
	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/btctypes"
)

type BtcProxy struct {
	Clients []btcrpc.Client
	Network btctypes.Network
}

func (proxy *BtcProxy) Blockinfo() btctypes.Network {
	panic("implement me")
}

func (proxy *BtcProxy) GetUTXOs(address btctypes.Addr, limit, confirmations int) ([]btctypes.UTXO, error) {
	errs := types.NewErrList(len(proxy.Clients))
	for i, client:= range proxy.Clients {
		utxos, err := client.GetUTXOs(address, limit, confirmations)
		if err != nil {
			errs[i] = err
			continue
		}
		return utxos, nil
	}
	return nil , errs
}

func (proxy *BtcProxy) Confirmations(txHash string) (int64, error) {
	errs := types.NewErrList(len(proxy.Clients))
	for i, client:= range proxy.Clients {
		confirmations, err := client.Confirmations(txHash)
		if err != nil {
			errs[i] = err
			continue
		}
		return confirmations, nil
	}
	return 0 , errs
}

func (proxy *BtcProxy) PublishTransaction(stx []byte) error {
	errs := types.NewErrList(len(proxy.Clients))
	for i, client:= range proxy.Clients {
		err := client.PublishTransaction(stx)
		if err != nil {
			errs[i] = err
			continue
		}
		return nil
	}
	return errs
}
