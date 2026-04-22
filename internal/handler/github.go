package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"pr_approved/internal/service"
)

type GitHubHandler struct {
	svc service.Service
}

func NewGitHubHandler(svc service.Service) *GitHubHandler {
	return &GitHubHandler{svc: svc}
}

// HandlePullRequest godoc
//
// @Summary      Perform actions on a GitHub pull request
// @Description  Approve, merge, or approve and merge a GitHub pull request.
// @Description  If no action is provided, defaults to "both" (approve + squash merge).
// @Description
// @Description  **Allowed actions:**
// @Description  - `approve` — Approves the pull request
// @Description  - `merge` — Squash merges the pull request
// @Description  - `both` — Approves and then squash merges (default)
// @Tags         pull-requests
// @Accept       json
// @Produce      json
// @Param        request body      GitHubRequest  true  "Pull request action payload"
// @Success      200     {object}  successResponse "Action completed successfully"
// @Failure      400     {object}  errorResponse   "Invalid URL, repository, or action"
// @Failure      500     {object}  errorResponse   "Failed to approve or merge the pull request"
// @Router       /git-hub [post]
func (h *GitHubHandler) HandlePullRequest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		slog.InfoContext(r.Context(), "request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}()

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		h.writeJSON(w, http.StatusInternalServerError, GitHubResponse{Message: err.Error(), Status: "error"})
		return
	}

	var req GitHubRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.writeJSON(w, http.StatusInternalServerError, GitHubResponse{Message: err.Error(), Status: "error"})
		return
	}

	owner, repo, prNumber, err := parseGitHubPRUrl(req.Url)
	if err != nil {
		h.writeJSON(w, http.StatusBadRequest, GitHubResponse{Message: err.Error(), Status: "error"})
		return
	}

	if !h.svc.IsValidOwner(owner) {
		h.writeJSON(w, http.StatusBadRequest, GitHubResponse{Message: service.ErrInvalidRepo.Error(), Status: "error"})
		return
	}

	action := strings.ToLower(req.Action)
	if action == "" {
		action = service.Both
	}

	if !h.svc.IsValidAction(action) {
		h.writeJSON(w, http.StatusBadRequest, GitHubResponse{Message: service.ErrInvalidAction.Error(), Status: "error"})
		return
	}

	if err := h.svc.HandlePR(r.Context(), owner, repo, prNumber, action); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrNoTokenForOwner) || errors.Is(err, service.ErrInvalidRepo) {
			status = http.StatusBadRequest
		}
		h.writeJSON(w, status, GitHubResponse{Message: err.Error(), Status: "error"})
		return
	}

	messages := map[string]string{
		service.Approve: "Pull request approved",
		service.Merge:   "Pull request merged",
		service.Both:    "Pull request approved and merged",
	}
	h.writeJSON(w, http.StatusOK, GitHubResponse{Message: messages[action], Status: "success"})
}

func parseGitHubPRUrl(rawUrl string) (owner, repo string, number int, err error) {
	parsed, err := url.Parse(rawUrl)
	if err != nil || parsed.Host != "github.com" {
		return "", "", 0, service.ErrInvalidUrl
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 4 || parts[2] != "pull" {
		return "", "", 0, service.ErrInvalidUrl
	}

	number, err = strconv.Atoi(parts[3])
	if err != nil {
		return "", "", 0, service.ErrInvalidUrl
	}

	return parts[0], parts[1], number, nil
}

func (h *GitHubHandler) writeJSON(w http.ResponseWriter, status int, response GitHubResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}
