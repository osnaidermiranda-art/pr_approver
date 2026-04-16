package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	docs "pr_approved/docs"

	httpSwagger "github.com/swaggo/http-swagger"

	"pr_approved/server"
)

// @title PR Approved API
// @version 1.0
// @description Service to approve and merge GitHub pull requests
// @BasePath /
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = "localhost:" + port
	}
	// Strip scheme if accidentally included (e.g. "https://example.com" -> "example.com")
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")

	docs.SwaggerInfo.Host = host
	docs.SwaggerInfo.Schemes = []string{"https", "http"}

	log.Println("Starting server on port " + port)
	srv := server.NewServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/git-hub", srv.HandlePullRequest)
	mux.HandleFunc("/", httpSwagger.WrapHandler)

	log.Fatal(http.ListenAndServe(":"+port, corsMiddleware(mux)))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
