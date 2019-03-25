package eth

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/renproject/mercury"
)

type ethereum struct {
	network   string
	tags      map[string]string
	initiated bool
}

func New(network string, tags map[string]string) mercury.BlockchainPlugin {
	return &ethereum{network, tags, true}
}

func (eth *ethereum) Init() error {
	return nil
}

func (eth *ethereum) Initiated() bool {
	return eth.initiated
}

// Handlers of the bitcoin blockchain
func (eth *ethereum) AddRoutes(r *mux.Router) {
	r.HandleFunc(eth.AddRoutePrefix(""), eth.jsonRPCHandler()).Queries("tag", "{tag}").Methods("POST")
	r.HandleFunc(eth.AddRoutePrefix(""), eth.jsonRPCHandler()).Methods("POST")
}

func (eth *ethereum) AddRoutePrefix(route string) string {
	return fmt.Sprintf("/%s-%s%s", "eth", eth.network, route)
}

func (eth *ethereum) jsonRPCHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := eth.tags[r.URL.Query().Get("tag")]
		if apiKey == "" {
			apiKey = eth.tags[""]
		}

		resp, err := http.Post(fmt.Sprintf("https://%s.infura.io/v3/%s", eth.network, apiKey), "application/json", r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("{ \"error\": \"%s\" }", err), resp.StatusCode)
			return
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("{ \"error\": \"%s\" }", err), resp.StatusCode)
			return
		}
		w.WriteHeader(resp.StatusCode)
		w.Write(data)
	}
}
