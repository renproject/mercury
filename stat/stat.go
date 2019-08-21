package stat

import (
	"sync"
	"time"
)

type Stat struct {
	// requestTimes is a map of hour -> method -> count
	requestTimes map[int]map[string]int
	// initTimes is a map of hour -> time
	initTimes  map[int]time.Time
	requestsMu *sync.Mutex
}

// day is the number of nanoseconds in a day
const day = 24 * time.Hour

func New() Stat {
	requestTimes := make(map[int]map[string]int)
	initTimes := make(map[int]time.Time)
	return Stat{
		requestTimes: requestTimes,
		initTimes:    initTimes,
		requestsMu:   &sync.Mutex{},
	}
}

func (stat *Stat) Get() map[string]int {
	t := time.Now()
	numRequests := make(map[string]int)
	stat.requestsMu.Lock()
	for hour, methodTimestamps := range stat.requestTimes {
		// only count the hour if it was within the past day
		if t.Sub(stat.initTimes[hour]) > day {
			continue
		}

		for method, count := range methodTimestamps {
			numRequests[method] += count
		}
	}
	stat.requestsMu.Unlock()
	return numRequests
}

func (stat *Stat) Insert(method string) {
	t := time.Now()
	stat.requestsMu.Lock()

	// initialise the secondary map if nil
	if stat.requestTimes[t.Hour()] == nil || t.Sub(stat.initTimes[t.Hour()]) > day {
		stat.requestTimes[t.Hour()] = make(map[string]int)
		stat.initTimes[t.Hour()] = t
	}

	stat.requestTimes[t.Hour()][method]++
	stat.requestsMu.Unlock()
}
