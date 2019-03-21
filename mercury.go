package mercury

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/time/rate"
)

type BlockchainPlugin interface {
	AddRoutes(router *mux.Router)
}

type Mercury interface {
	Run()
}

type server struct {
	port    string
	plugins []BlockchainPlugin
}

// New mercury http server
func New(port string, plugins ...BlockchainPlugin) Mercury {
	return &server{
		port:    port,
		plugins: plugins,
	}
}

func (server *server) Run() {
	limiter := rate.NewLimiter(3, 20)
	r := mux.NewRouter()
	for _, plugin := range server.plugins {
		plugin.AddRoutes(r)
	}
	r.Use(func(handler http.Handler) http.Handler {
		return rateLimit(limiter, handler)
	})
	r.Use(recoveryHandler)
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST"},
	}).Handler(r)
	log.Printf("Listening on port %v...", server.port)
	http.ListenAndServe(fmt.Sprintf(":%v", server.port), handler)
}

func rateLimit(limiter *rate.Limiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if limiter.Allow() {
			next.ServeHTTP(w, r)
			return
		}
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("too many requests"))
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
