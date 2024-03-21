package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type apiConfig struct {
	fileserverHits int
}

func (cfg *apiConfig) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits += 1
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) printFileserverHits() string {
	return fmt.Sprintf("Hits: %d", cfg.fileserverHits)
}

func (cfg *apiConfig) resetFileserverHits() {
	cfg.fileserverHits = 0
}

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	type parameters struct {
		Body string `json:"body"`
	}

	type returnSuccess struct {
		Valid bool `json:"valid"`
	}

	type returnError struct {
		Error string `json:"error"`
	}

	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		respBody := returnError{
			Error: "Something went wrong",
		}
		dat, err := json.Marshal(respBody)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(dat)
		return
	}

	if len(params.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest)
		respBody := returnError{
			Error: "Chirp is too long",
		}
		dat, err := json.Marshal(respBody)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(dat)
		return
	}

	w.WriteHeader(http.StatusOK)
	respBody := returnSuccess{
		Valid: true,
	}
	dat, err := json.Marshal(respBody)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(dat)
	return
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	metrics := apiConfig{fileserverHits: 0}
	mux := http.NewServeMux()
	fileServerHandler := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", metrics.metricsMiddleware(fileServerHandler))
	mux.HandleFunc("/admin/metrics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset+utf-8")
		w.WriteHeader(http.StatusOK)
		response := fmt.Sprintf(`<html>
					<body>
						<h1>Welcome, Chirpy Admin</h1>
						<p>Chirpy has been visited %d times!</p>
					</body>
				</html>
		`, metrics.fileserverHits)
		w.Write([]byte(response))
	})
	mux.HandleFunc("/api/reset", func(w http.ResponseWriter, r *http.Request) {
		metrics.resetFileserverHits()
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/api/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/api/validate_chirp", validateChirpHandler)

	corsMux := corsMiddleware(mux)

	server := &http.Server{
		Addr:    ":8080",
		Handler: corsMux,
	}

	server.ListenAndServe()
}
