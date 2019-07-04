package btcaccount

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Speed indicates the tier of speed that the transaction falls under while writing to the blockchain.
type Speed uint8

// TxExecutionSpeed values.
const (
	Nil = Speed(iota)
	Slow
	Standard
	Fast
)

// BtcGasStation retrieves the recommended tx fee from `bitcoinfees.earn.com`. It cached the result to avoid hitting the
// rate limiting of the API. It's safe for using concurrently.
type BtcGasStation interface {
	GasRequired(ctx context.Context, speed Speed) int64
}

type btcGasStation struct {
	mu            *sync.RWMutex
	logger        logrus.FieldLogger
	fees          map[Speed]int64
	lastUpdate    time.Time
	minUpdateTime time.Duration
}

// NewBtcGasStation returns a new BtcGasStation
func NewBtcGasStation(logger logrus.FieldLogger, minUpdateTime time.Duration) BtcGasStation {
	return &btcGasStation{
		mu:            new(sync.RWMutex),
		logger:        logger,
		fees:          map[Speed]int64{},
		lastUpdate:    time.Time{},
		minUpdateTime: minUpdateTime,
	}
}

func (btc btcGasStation) GasRequired(ctx context.Context, speed Speed) int64 {
	btc.mu.Lock()
	defer btc.mu.Unlock()

	if time.Now().After(btc.lastUpdate.Add(btc.minUpdateTime)) {
		if err := btc.gasRequired(ctx); err != nil {
			btc.logger.Errorf("cannot get recommended fee from bitcoinfees.earn.com, err = %v", err)
		}
	}

	return btc.fees[speed]
}

func (btc *btcGasStation) gasRequired(ctx context.Context) error {
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
	btc.fees[Fast] = fee.Fast
	btc.fees[Standard] = fee.Standard
	btc.fees[Slow] = fee.Slow
	btc.lastUpdate = time.Now()
	return nil
}
