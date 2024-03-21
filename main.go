package main

import (
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

	corsMux := corsMiddleware(mux)

	server := &http.Server{
		Addr:    ":8080",
		Handler: corsMux,
	}

	server.ListenAndServe()
}
