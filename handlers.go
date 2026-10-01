package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type errorMsg struct {
	Error string `json:"error"`
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

func (cfg *apiConfig) handleFileserverHits(r http.ResponseWriter, req *http.Request) {
	r.Header().Add("Content-Type", "text/html")
	r.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handleReset(r http.ResponseWriter, req *http.Request) {
	if cfg.Platform != "dev" {
		r.WriteHeader(403)
		return
	}
	cfg.fileserverHits.Swap(0)
	r.Write([]byte("Hits reset."))
	cfg.db.DeleteAllUsers(req.Context())
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}
	type response struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}
	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithJSON(w, 400, errorMsg{Error: "Bad Request"})
		return
	}

	user, err := cfg.db.CreateUser(req.Context(), params.Email)
	if err != nil {
		respondWithJSON(w, 400, errorMsg{Error: "Bad Request"})
		return
	}

	respondWithJSON(w, 201, response{Id: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Email: user.Email})
}
