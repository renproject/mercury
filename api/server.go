package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/renproject/mercury/stat"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

type BlockchainApi interface {
	AddHandler(r *mux.Router, s *stat.Stat)
}

// DefaultMaxHeaderBytes is the maximum permitted size of the headers in an HTTP request.
const DefaultMaxHeaderBytes = 1 << 10 // 1 KB

type Server struct {
	apis   []BlockchainApi
	port   string
	logger logrus.FieldLogger
	stat   *stat.Stat
}

// NewServer returns a server which supports the given blockchain APIs.
func NewServer(logger logrus.FieldLogger, port string, apis ...BlockchainApi) *Server {
	s := stat.NewStat()
	return &Server{
		apis:   apis,
		port:   port,
		logger: logger,
		stat:   &s,
	}
}

// Run starts the server.
func (server *Server) Run() {
	// Add handlers for each blockchain.
	r := mux.NewRouter().StrictSlash(true)
	for _, api := range server.apis {
		api.AddHandler(r, server.stat)
	}
	r.HandleFunc("/health", server.health()).Methods("GET")
	r.HandleFunc("/stats", server.stats()).Methods("GET")

	// Use recovery handler and provide cross-origin support.
	r.Use(server.recoveryHandler)
	handler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST"},
	}).Handler(r)

	// Set-up request timeout and header size limit for the server.
	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%v", server.port),
		Handler:           handler,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 3 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       1 * time.Minute,
		MaxHeaderBytes:    DefaultMaxHeaderBytes,
	}

	// Start running the server.
	server.logger.Infof("mercury listening on 0.0.0.0:%v...", server.port)
	httpServer.ListenAndServe()
}

func (server *Server) health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func (server *Server) stats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := server.stat.Get()
		json.NewEncoder(w).Encode(data)
		w.Header().Set("Content-Type", "application/json")
	}
}

func (server *Server) recoveryHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				server.logger.Errorf("recovering from: %v", r)
				http.Error(w, fmt.Sprintf("recovery from: %v", r), http.StatusInternalServerError)
			}
		}()

		h.ServeHTTP(w, r)
	})
}
