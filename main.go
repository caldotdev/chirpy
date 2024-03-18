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
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset+utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(metrics.printFileserverHits()))
	})
	mux.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		metrics.resetFileserverHits()
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
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
