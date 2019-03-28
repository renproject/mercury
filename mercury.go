package mercury

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/republicprotocol/co-go"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

type BlockchainPlugin interface {
	Init() error
	Initiated() bool
	AddRoutes(router *mux.Router)
}

type Mercury interface {
	Run()
}

type server struct {
	port    string
	logger  logrus.FieldLogger
	plugins []BlockchainPlugin
}

// New mercury http server
func New(port string, logger logrus.FieldLogger, plugins ...BlockchainPlugin) Mercury {
	go co.ParForAll(plugins, func(i int) {
		if err := plugins[i].Init(); err != nil {
			logger.Error(err)
		}
	})
	return &server{
		port:    port,
		logger:  logger,
		plugins: plugins,
	}
}

func (server *server) Run() {
	r := mux.NewRouter()
	for _, plugin := range server.plugins {
		plugin.AddRoutes(r)
	}
	r.Use(recoveryHandler)
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST"},
	}).Handler(r)
	log.Printf("Listening on port %v...", server.port)
	http.ListenAndServe(fmt.Sprintf(":%v", server.port), handler)
}

func isInitiated(plugin BlockchainPlugin, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if plugin.Initiated() {
			next.ServeHTTP(w, r)
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("service unavailable"))
	})
}

func recoveryHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				http.Error(w, fmt.Sprintf("recovery from: %v", r), http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, r)
	})
}
