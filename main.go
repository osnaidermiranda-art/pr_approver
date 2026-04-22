package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	_ "pr_approved/docs"
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
// @host localhost:8080
// @BasePath /
func main() {
	_ = godotenv.Load()

	clients := buildClients()
	validOwners := buildValidOwners()

	svc := service.NewGitHubService(clients, validOwners)
	gh := handler.NewGitHubHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /git-hub", gh.HandlePullRequest)
	mux.HandleFunc("/", httpSwagger.WrapHandler)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("server starting", "addr", ":8080")

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
		owners[strings.TrimSpace(r)] = true
	}
	return owners
}
