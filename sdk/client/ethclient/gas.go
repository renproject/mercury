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
	GasRequired(ctx context.Context, speed types.TxSpeed) (ethtypes.Amount, error)
	Initialized() bool
}

type ethGasStation struct {
	mu            *sync.RWMutex
	logger        logrus.FieldLogger
	fees          map[types.TxSpeed]ethtypes.Amount
	lastUpdate    time.Time
	minUpdateTime time.Duration
}

// NewEthGasStation returns a new EthGasStation
func NewEthGasStation(logger logrus.FieldLogger, minUpdateTime time.Duration) EthGasStation {
	return &ethGasStation{
		mu:            new(sync.RWMutex),
		logger:        logger,
		fees:          map[types.TxSpeed]ethtypes.Amount{},
		lastUpdate:    time.Time{},
		minUpdateTime: minUpdateTime,
	}
}

func (eth ethGasStation) Initialized() bool {
	return !eth.lastUpdate.IsZero()
}

// GasRequired returns the recommended gas price
func (eth ethGasStation) GasRequired(ctx context.Context, speed types.TxSpeed) (amount ethtypes.Amount, err error) {
	eth.mu.Lock()
	defer eth.mu.Unlock()

	if time.Now().After(eth.lastUpdate.Add(eth.minUpdateTime)) {
		if err = eth.gasRequired(ctx); err != nil {
			eth.logger.Errorf("cannot get recommended fee from ethgasstation.info, err = %v", err)
		}
	}

	// Return an error if we were never able to fetch gas prices
	if !eth.Initialized() {
		return ethtypes.Amount{}, fmt.Errorf("failed to fetch gas information: %v", err)
	}

	return eth.fees[speed], nil
}

func (eth *ethGasStation) gasRequired(ctx context.Context) error {
	request, err := http.NewRequest("GET", "https://ethgasstation.info/json/ethgasAPI.json", nil)
	if err != nil {
		return fmt.Errorf("cannot build request to ethGasStation = %v", err)
	}
	request.Header.Set("Content-Type", "application/json")

	res, err := (&http.Client{}).Do(request)
	if err != nil {
		return fmt.Errorf("cannot connect to ethGasStationAPI = %v", err)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %v from ethGasStation", res.StatusCode)
	}

	data := struct {
		Slow     float64 `json:"safeLow"`
		Standard float64 `json:"average"`
		Fast     float64 `json:"fast"`
	}{}
	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		return fmt.Errorf("cannot decode response body from ethGasStation = %v", err)
	}

	eth.fees[types.Fast] = ethtypes.Gwei(uint64(data.Fast))
	eth.fees[types.Standard] = ethtypes.Gwei(uint64(data.Standard))
	eth.fees[types.Slow] = ethtypes.Gwei(uint64(data.Slow))
	eth.lastUpdate = time.Now()
	return nil
}
