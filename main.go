package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

func main() {
	apiCfg := apiConfig{}

	serveMux := http.NewServeMux()
	serveMux.HandleFunc("GET /api/healthz", handleReadiness)
	serveMux.HandleFunc("GET /admin/metrics", apiCfg.handleFileserverHits)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.handleReset)
	serveMux.HandleFunc("POST /api/validate_chirp", handleValidateChirp)
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

func handleValidateChirp(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type errorMsg struct {
		Error string `json:"error"`
	}
	type cleanedMsg struct {
		CleanedBody string `json:"cleaned_body"`
	}
	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		errMsg := errorMsg{Error: "Something went wrong"}
		respondWithJSON(w, 500, errMsg)
		return
	}
	if len(params.Body) > 140 {
		errMsg := errorMsg{Error: "Chirp is too long"}
		respondWithJSON(w, 400, errMsg)
		return
	}

	words := strings.Split(params.Body, " ")
	for i, word := range words {
		word := strings.ToLower(word)
		if word == "kerfuffle" || word == "sharbert" || word == "fornax" {
			words[i] = "****"
		}
	}

	cleaned := cleanedMsg{CleanedBody: strings.Join(words, " ")}
	respondWithJSON(w, 200, cleaned)
}

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")

	data, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("{\"error\": \"Something went wrong!\"}"))
		fmt.Println(err)
		return
	}

	w.WriteHeader(code)
	w.Write(data)
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
	r.Header().Add("Content-Type", "text/html")
	r.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handleReset(r http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Swap(0)
	r.Write([]byte("Hits reset."))
}
