package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	docs "pr_approved/docs"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"

	"pr_approved/internal/ghclient"
	"pr_approved/internal/handler"
	"pr_approved/internal/service"
)

const githubTokenPrefix = "GITHUB_TOKEN_"

// @title PR Approved API
// @version 1.0
// @description Service to approve and merge GitHub pull requests
// @BasePath /
func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = "localhost:" + port
	}

	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "https://")

	docs.SwaggerInfo.Host = host
	docs.SwaggerInfo.Schemes = []string{"https", "http"}

	clients := buildClients()
	validOwners := buildValidOwners()

	svc := service.NewGitHubService(clients, validOwners)
	gh := handler.NewGitHubHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /git-hub", gh.HandlePullRequest)
	mux.HandleFunc("/", httpSwagger.WrapHandler)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMiddleware(mux),
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("server starting", "addr", host)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "err", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}

func buildClients() map[string]ghclient.Client {
	clients := make(map[string]ghclient.Client)

	for _, env := range os.Environ() {
		key, token, ok := strings.Cut(env, "=")
		if !ok || !strings.HasPrefix(key, githubTokenPrefix) || token == "" {
			continue
		}
		owner := strings.ToLower(strings.ReplaceAll(strings.TrimPrefix(key, githubTokenPrefix), "_", "-"))
		clients[owner] = ghclient.New(token)
	}

	if len(clients) == 0 {
		slog.Error("no GitHub tokens found", "prefix", githubTokenPrefix)
		os.Exit(1)
	}

	return clients
}

func buildValidOwners() map[string]bool {
	raw := os.Getenv("VALID_OWNERS")
	if raw == "" {
		slog.Error("VALID_OWNERS environment variable is required")
		os.Exit(1)
	}
	owners := make(map[string]bool)
	for r := range strings.SplitSeq(raw, ",") {
		owners[strings.ToLower(strings.TrimSpace(r))] = true
	}
	return owners
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}