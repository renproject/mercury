package ethclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/renproject/mercury/types"
	"github.com/renproject/mercury/types/ethtypes"
	"github.com/sirupsen/logrus"
)

// EthGasStation retrieves the recommended tx fee from `bitcoinfees.earn.com`. It cached the result to avoid hitting the
// rate limiting of the API. It's safe for using concurrently.
type EthGasStation interface {
	GasRequired(ctx context.Context, speed types.TxSpeed) int64
	CalculateGasAmount(ctx context.Context, speed types.TxSpeed, txSizeInBytes int) ethtypes.Amount
}

type ethGasStation struct {
	mu            *sync.RWMutex
	logger        logrus.FieldLogger
	fees          map[types.TxSpeed]int64
	lastUpdate    time.Time
	minUpdateTime time.Duration
}

// NewEthGasStation returns a new EthGasStation
func NewEthGasStation(logger logrus.FieldLogger, minUpdateTime time.Duration) EthGasStation {
	return &ethGasStation{
		mu:            new(sync.RWMutex),
		logger:        logger,
		fees:          map[types.TxSpeed]int64{},
		lastUpdate:    time.Time{},
		minUpdateTime: minUpdateTime,
	}
}

func (eth ethGasStation) GasRequired(ctx context.Context, speed types.TxSpeed) int64 {
	eth.mu.Lock()
	defer eth.mu.Unlock()

	if time.Now().After(eth.lastUpdate.Add(eth.minUpdateTime)) {
		if err := eth.gasRequired(ctx); err != nil {
			eth.logger.Errorf("cannot get recommended fee from bitcoinfees.earn.com, err = %v", err)
		}
	}

	return eth.fees[speed]
}

func (eth ethGasStation) CalculateGasAmount(ctx context.Context, speed types.TxSpeed, txSizeInBytes int) ethtypes.Amount {
	gasRequired := eth.GasRequired(ctx, speed) // in sats/byte
	gasInSats := gasRequired * int64(txSizeInBytes)
	return ethtypes.Amount(gasInSats)
}

func (eth *ethGasStation) gasRequired(ctx context.Context) error {
	// FIXME: Use context for http request timeout
	response, err := http.Get("https://bitcoinfees.earn.com/api/v1/fees/recommended")
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %v", response.StatusCode)
	}

	var fee = struct {
		Fast     int64 `json:"fastestFee"`
		Standard int64 `json:"halfHourFee"`
		Slow     int64 `json:"hourFee"`
	}{}
	if err := json.NewDecoder(response.Body).Decode(&fee); err != nil {
		return err
	}
	eth.fees[types.Fast] = fee.Fast
	eth.fees[types.Standard] = fee.Standard
	eth.fees[types.Slow] = fee.Slow
	eth.lastUpdate = time.Now()
	return nil
}
