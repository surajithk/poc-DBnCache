package main

import (
	"fmt"
	"time"
	"net/http"
	cache "github.com/NYTimes/mercury-poc/cache"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

const (
	redisHost = "10.180.141.36"
)

func main() {
	defer fmt.Println("poc exited")

	start := time.Now()

	c := cache.NewCache(redisHost)

	r := mux.NewRouter()
	r.HandleFunc("/healthz", handleHealth)
	r.HandleFunc("/read/cache", handleCache)
	r.HandleFunc("/read/db", handleDb)
	http.Handle("/", r)

	fmt.Printf("Send Compiler is up and running in %v!\n", time.Since(start).Seconds())
}

func (s *Server) handleHealth() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		_, _ = w.Write([]byte("OK"))
	}
}

func (s *Server) handleCache() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		_, _ = w.Write([]byte("OK"))
	}
}

func (s *Server) handleDb() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		_, _ = w.Write([]byte("OK"))
	}
}
