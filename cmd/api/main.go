package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"linkforge/internal/handlers"
	"linkforge/internal/store"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := store.NewPostgres(ctx)
	if err != nil {
		log.Fatalf("database initialization failed: %v", err)
	}
	defer db.Close()

	linkHandler := handlers.NewLinkHandler(db)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
	})
	limiter := handlers.NewRateLimiter(10, time.Minute)
	mux.Handle(
		"POST /shorten",
		limiter.Middleware(http.HandlerFunc(linkHandler.Shorten)),
	)
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.NotFound(w, r)
        return
    }

    http.ServeFile(w, r, "/app/web/index.html")
    })

    mux.HandleFunc("GET /style.css", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "/app/web/style.css")
    })

    mux.HandleFunc("GET /app.js", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "/app/web/app.js")
    })
	mux.HandleFunc("GET /{code}", linkHandler.Redirect)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("LinkForge listening on :8080")
	log.Fatal(server.ListenAndServe())
}
