package main

import (
	"encoding/json"
	"net/http"
	"os"
)

// Заглушка до ARCH-03: минимальный HTTP-сервер с healthcheck.
// Путь cmd/server/main.go фиксирован контрактом DOL-02.
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		panic(err)
	}
}
