package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bennetseidelic/chirpy/internal/database"
	"github.com/google/uuid"
)

type errorMsg struct {
	Error string `json:"error"`
}

type chirpResponse struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

func handleReadiness(responseWriter http.ResponseWriter, req *http.Request) {
	header := responseWriter.Header()
	header.Add("Content-Type", "text/plain; charset=utf-8")
	responseWriter.WriteHeader(200)
	responseWriter.Write([]byte("OK"))
}

func isValidChirp(chirp string) bool {
	return len(chirp) <= 140
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

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}
	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithJSON(w, 400, errorMsg{Error: "Bad Request"})
		fmt.Println(err)
		return
	}
	if !isValidChirp(params.Body) {
		respondWithJSON(w, 400, errorMsg{Error: "Bad Request"})
		fmt.Println(err)
		return
	}

	chirp, err := cfg.db.CreateChirp(req.Context(), database.CreateChirpParams{Body: params.Body, UserID: params.UserId})
	if err != nil {
		respondWithJSON(w, 400, errorMsg{Error: "Bad Request"})
		fmt.Println(err)
		return
	}

	respondWithJSON(w, 201, chirpResponse{Id: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, Body: chirp.Body, UserId: chirp.UserID})
}

func (cfg *apiConfig) handleGetChirps(w http.ResponseWriter, req *http.Request) {

	chirps, err := cfg.db.GetAllChirps(req.Context())
	if err != nil {
		respondWithJSON(w, 500, errorMsg{Error: "Something went wrong"})
		return
	}
	chirpResponses := []chirpResponse{}
	for _, chirp := range chirps {
		chirpResponses = append(chirpResponses, chirpResponse{
			Id:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserId:    chirp.UserID,
		})
	}

	respondWithJSON(w, 200, chirpResponses)
}

func (cfg *apiConfig) handleGetChirp(w http.ResponseWriter, req *http.Request) {
	chirpId := req.PathValue("chirpID")
	chirpUuid, err := uuid.Parse(chirpId)
	if err != nil {
		fmt.Println(err)
		return
	}
	chirp, err := cfg.db.GetChirp(req.Context(), chirpUuid)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(404)
		return
	}

	respondWithJSON(w, 200, chirpResponse{Id: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, Body: chirp.Body, UserId: chirp.UserID})
}
