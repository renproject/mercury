package eth

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type RelayRequest struct {
	Address string   `json:"address"`
	FnName  string   `json:"fnName"`
	Data    []string `json:"data"`
}

type RelayResponse struct {
	TxHash string `json:"txHash"`
}

func (eth *ethereum) Relay(req RelayRequest) (RelayResponse, error) {
	data := make([][]byte, len(req.Data))
	var err error
	for i := range data {
		if len(req.Data[i]) >= 2 && req.Data[i][:2] == "0x" {
			req.Data[i] = req.Data[i][2:]
		}

		data[i], err = hex.DecodeString(req.Data[i])
		if err != nil {
			return RelayResponse{}, err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	tx, err := eth.account.ContractTransact(ctx, common.HexToAddress(req.Address), req.FnName, data...)
	if err != nil {
		return RelayResponse{}, err
	}

	return RelayResponse{tx.Hash().String()}, nil
}
