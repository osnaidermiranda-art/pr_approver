package main

import (
	"log"
	"net/http"
	"os"
	_ "pr_approved/docs"

	httpSwagger "github.com/swaggo/http-swagger"

	"pr_approved/server"
)

// @title PR Approved API
// @version 1.0
// @description Service to approve and merge GitHub pull requests
// @host localhost:8080
// @BasePath /
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Starting server on port " + port)
	srv := server.NewServer()

	http.HandleFunc("/git-hub", srv.HandlePullRequest)
	http.HandleFunc("/", httpSwagger.WrapHandler)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
