package main

import (
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
)

func main() {
	apiCfg := apiConfig{}

	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/healthz", handleReadiness)
	serveMux.HandleFunc("/metrics", apiCfg.handleFileserverHits)
	serveMux.HandleFunc("/reset", apiCfg.handleReset)
	serveMux.Handle("/app/", http.StripPrefix("/app/", apiCfg.middlewareMetricsInc(http.FileServer((http.Dir("."))))))

	server := http.Server{Handler: serveMux, Addr: ":8080"}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func handleReadiness(responseWriter http.ResponseWriter, req *http.Request) {
	header := responseWriter.Header()
	header.Add("Content-Type", "text/plain; charset=utf-8")
	responseWriter.WriteHeader(200)
	responseWriter.Write([]byte("OK"))
}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleFileserverHits(r http.ResponseWriter, req *http.Request) {
	r.Write([]byte("Hits: " + strconv.FormatInt(int64(cfg.fileserverHits.Load()), 10)))
}

func (cfg *apiConfig) handleReset(r http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Swap(0)
	r.Write([]byte("Hits reset."))
}
