package mercury

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/republicprotocol/co-go"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

type BlockchainPlugin interface {
	Init() error
	Initiated() bool
	Health() bool
	Prefix() string
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
	r.Use(isInitiatedMiddleware(server.plugins))
	r.HandleFunc("/health", server.getHealth()).Methods("GET")
	r.Use(server.recoveryHandler)
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST"},
	}).Handler(r)
	log.Printf("Listening on port %v...", server.port)
	http.ListenAndServe(fmt.Sprintf(":%v", server.port), handler)
}

func (server *server) getHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		healthMap := map[string]bool{}
		for _, plugin := range server.plugins {
			healthMap[plugin.Prefix()] = plugin.Health()
		}
		if err := json.NewEncoder(w).Encode(healthMap); err != nil {
			server.logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func isInitiatedMiddleware(plugins []BlockchainPlugin) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, plugin := range plugins {
				if plugin.Prefix() == strings.Split(r.URL.Path, "/")[1] {
					if !plugin.Initiated() {
						w.WriteHeader(http.StatusServiceUnavailable)
						w.Write([]byte("service unavailable"))
						return
					}

				}
			}
			h.ServeHTTP(w, r)
		})
	}
}

func (server *server) recoveryHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				server.logger.Error(r)
				http.Error(w, fmt.Sprintf("recovery from: %v", r), http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, r)
	})
}
