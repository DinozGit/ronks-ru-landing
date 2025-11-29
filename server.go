package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

const (
	email = "test_ins@mappex.ru"
	ui    = "00000000-0000-0000-0000-000000000000"
	ver   = "3.0.1"
	apiURL = "https://fapi.iisis.ru/fapi/v2/analogList"
	addr   = ":8008" // listen on all interfaces (needed for Docker)
)

type errorResponse struct {
	Error string `json:"error"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/analog", handleAnalog)
	mux.HandleFunc("/", handleRoot)

	absIndex, _ := filepath.Abs("index.html")
	log.Printf("✅ Go-сервер запущен: http://%s", addr)
	log.Printf("📁 Выдаёт HTML-файл: %s", absIndex)

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func handleAnalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "Method not allowed"})
		return
	}

	partNumber := r.URL.Query().Get("n")
	if partNumber == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "Missing 'n' parameter"})
		return
	}

	u, err := url.Parse(apiURL)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Internal URL parse error"})
		return
	}

	q := u.Query()
	q.Set("n", partNumber)
	q.Set("email", email)
	q.Set("ui", ui)
	q.Set("ver", ver)
	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(u.String())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "API error: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("write response error: %v", err)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	// Serve static assets if the requested path exists on disk
	if r.URL.Path != "/" && r.URL.Path != "" {
		path := "." + r.URL.Path // e.g. /css/bootstrap.css -> ./css/bootstrap.css
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			http.ServeFile(w, r, path)
			return
		}
	}

	// Fallback: always serve index.html for SPA-style routing
	if _, err := os.Stat("index.html"); err != nil {
		http.Error(w, "index.html not found", http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, "index.html")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("json encode error: %v", err)
	}
}
