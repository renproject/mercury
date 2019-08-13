package stat

import (
	"sync"
	"time"
)

type Stat struct {
	// requests is a
	requestTimes map[string]map[int]int
	requestsMu   *sync.Mutex
}

func NewStat() Stat {
	requestTimes := make(map[string]map[int]int)
	return Stat{
		requestTimes: requestTimes,
		requestsMu:   &sync.Mutex{},
	}
}

func (stat *Stat) Get() map[string]int {
	numRequests := make(map[string]int)
	stat.requestsMu.Lock()
	for method, hourTimestamps := range stat.requestTimes {
		total := 0
		for _, count := range hourTimestamps {
			total += count
		}
		numRequests[method] = total
	}
	stat.requestsMu.Unlock()
	return numRequests
}

func (stat *Stat) Insert(method string) {
	t := time.Now()
	stat.requestsMu.Lock()
	stat.requestTimes[method][t.Hour()]++
	stat.requestsMu.Unlock()
}
