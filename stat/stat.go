package stat

type Stat struct {
	// requests is a
	requests map[string][]int64
	requestsMu
}

func NewStat() Stat {
	requestTimes := make(map[string][]int64)
	return Stat{requestTimes}
}

func (stat *Stat) Get() map[string]int {
	numRequests := make(map[string]int)
	for method, timestamps := range stat.requestTimes {

	}
}

func (stat *Stat) Insert(method string) {
	stat.numRequests[method]++
}
