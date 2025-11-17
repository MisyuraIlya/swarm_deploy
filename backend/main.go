package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type App struct {
	db *sql.DB
}

type Item struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

type createItemRequest struct {
	Title string `json:"title"`
}

func main() {
	dsn := buildDSNFromEnv()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("failed to open DB: %v", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("failed to ping DB: %v", err)
	}

	if err := migrate(db); err != nil {
		log.Fatalf("failed to run migrate: %v", err)
	}

	app := &App{db: db}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", app.handleHealth)
	mux.HandleFunc("/api/items", app.handleItems)

	// CORS for frontend on different port
	handler := withCORS(mux)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("backend listening on :8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func buildDSNFromEnv() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "app")
	password := getEnv("DB_PASSWORD", "secret")
	dbName := getEnv("DB_NAME", "appdb")
	sslmode := getEnv("DB_SSLMODE", "disable")

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		url.QueryEscape(user),
		url.QueryEscape(password),
		host,
		port,
		dbName,
		sslmode,
	)
}

func getEnv(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

func migrate(db *sql.DB) error {
	const q = `
CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);`
	_, err := db.Exec(q)
	return err
}

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	if err := a.db.PingContext(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "down"})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (a *App) handleItems(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.listItems(w, r)
	case http.MethodPost:
		a.createItem(w, r)
	case http.MethodOptions:
		// handled by CORS middleware, but OK to return 200 here
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *App) createItem(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req createItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	var item Item
	err := a.db.QueryRowContext(
		r.Context(),
		`INSERT INTO items (title) VALUES ($1) RETURNING id, title, created_at`,
		req.Title,
	).Scan(&item.ID, &item.Title, &item.CreatedAt)

	if err != nil {
		log.Printf("failed to insert item: %v", err)
		http.Error(w, "failed to create item", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(item)
}

func (a *App) listItems(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.QueryContext(
		r.Context(),
		`SELECT id, title, created_at FROM items ORDER BY created_at DESC`,
	)
	if err != nil {
		log.Printf("failed to query items: %v", err)
		http.Error(w, "failed to load items", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := make([]Item, 0, 16)
	for rows.Next() {
		var it Item
		if err := rows.Scan(&it.ID, &it.Title, &it.CreatedAt); err != nil {
			log.Printf("failed to scan item: %v", err)
			http.Error(w, "failed to load items", http.StatusInternalServerError)
			return
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		log.Printf("rows error: %v", err)
		http.Error(w, "failed to load items", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(items)
}

// very simple CORS for demo (ok for learning, tighten in real prod)
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For learning: allow everything. In real app, restrict origin.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
