package server

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v84/github"
)

const (
	APPROVE = "approve"
	MERGE   = "merge"
	BOTH    = "both"
)

var validActions = map[string]bool{
	APPROVE: true,
	MERGE:   true,
	BOTH:    true,
}

var (
	ErrInvalidUrl        = errors.New("Invalid URL")
	ErrInvalidRepo       = errors.New("Invalid repository")
	ErrInvalidAction     = errors.New("Invalid action")
	ErrFailedToApprovePr = errors.New("Failed to approve pull request")
	ErrFailedToMergePr   = errors.New("Failed to merge pull request")
)

type GitHubRequest struct {
	Url    string `json:"url" example:"https://github.com/G97-TECH-MKT/my-repo/pull/42"`
	Action string `json:"action" example:"both" enums:"approve,merge,both"`
} // @name GitHubRequest

type GitHubResponse struct {
	Message string `json:"message"`
	Status  string `json:"status" enums:"success,error"`
}

type successResponse struct {
	Message string `json:"message" example:"Pull request approved and merged"`
	Status  string `json:"status" example:"success"`
} // @name SuccessResponse

type errorResponse struct {
	Message string `json:"message" example:"Invalid action"`
	Status  string `json:"status" example:"error"`
} // @name ErrorResponse

type Server struct {
	client     *github.Client
	validRepos map[string]bool
}

func NewServer() *Server {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN environment variable is required")
	}

	rawRepos := os.Getenv("VALID_REPOS")
	if rawRepos == "" {
		log.Fatal("VALID_REPOS environment variable is required")
	}

	validRepos := make(map[string]bool)
	for _, r := range strings.Split(rawRepos, ",") {
		validRepos[strings.TrimSpace(r)] = true
	}

	return &Server{
		client:     github.NewClient(nil).WithAuthToken(token),
		validRepos: validRepos,
	}
}

// HandlePullRequest godoc
//
// @Summary      Perform actions on a GitHub pull request
// @Description  Approve, merge, or approve and merge a GitHub pull request from the G97-TECH-MKT organization.
// @Description  If no action is provided, defaults to "both" (approve + squash merge).
// @Description
// @Description  **Allowed actions:**
// @Description  - `approve` — Approves the pull request
// @Description  - `merge` — Squash merges the pull request
// @Description  - `both` — Approves and then squash merges (default)
// @Description
// @Description  **URL format:** `https://github.com/G97-TECH-MKT/{repo}/pull/{number}`
// @Tags         pull-requests
// @Accept       json
// @Produce      json
// @Param        request body      GitHubRequest  true  "Pull request action payload"
// @Success      200     {object}  successResponse "Action completed successfully"
// @Failure      400     {object}  errorResponse   "Invalid URL, repository, or action"
// @Failure      405     {object}  errorResponse   "Method not allowed (only POST)"
// @Failure      500     {object}  errorResponse   "Failed to approve or merge the pull request"
// @Router       /git-hub [post]
func (s *Server) HandlePullRequest(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr)

	if r.Method != http.MethodPost {
		s.writeJSON(w, http.StatusMethodNotAllowed, GitHubResponse{
			Message: "Method not allowed",
			Status:  "error",
		})
		return
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, GitHubResponse{
			Message: err.Error(),
			Status:  "error",
		})
		return
	}

	bodyRequest := GitHubRequest{}
	err = json.Unmarshal(body, &bodyRequest)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, GitHubResponse{
			Message: err.Error(),
			Status:  "error",
		})
		return
	}

	owner, repo, prNumber, err := s.parseGitHubPRUrl(bodyRequest.Url)
	if err != nil {
		s.writeJSON(w, http.StatusBadRequest, GitHubResponse{
			Message: err.Error(),
			Status:  "error",
		})
		return
	}

	action := strings.ToLower(bodyRequest.Action)
	if action == "" {
		action = BOTH
	}

	if !validActions[action] {
		s.writeJSON(w, http.StatusBadRequest, GitHubResponse{
			Message: ErrInvalidAction.Error(),
			Status:  "error",
		})
		return
	}

	if action == APPROVE || action == BOTH {
		err = s.approvePullRequest(r, owner, repo, prNumber)
		if err != nil {
			s.writeJSON(w, http.StatusInternalServerError, GitHubResponse{
				Message: ErrFailedToApprovePr.Error(),
				Status:  "error",
			})
			return
		}
	}

	if action == MERGE || action == BOTH {
		err = s.mergePullRequest(r, owner, repo, prNumber)
		if err != nil {
			s.writeJSON(w, http.StatusInternalServerError, GitHubResponse{
				Message: ErrFailedToMergePr.Error(),
				Status:  "error",
			})
			return
		}
	}

	messages := map[string]string{
		APPROVE: "Pull request approved",
		MERGE:   "Pull request merged",
		BOTH:    "Pull request approved and merged",
	}
	s.writeJSON(w, http.StatusOK, GitHubResponse{
		Message: messages[action],
		Status:  "success",
	})
}

// Extracts owner, repo, and PR number from a GitHub PR URL
// e.g. https://github.com/G97-TECH-MKT/my-repo/pull/42
func (s *Server) parseGitHubPRUrl(rawUrl string) (owner, repo string, number int, err error) {
	parsed, err := url.Parse(rawUrl)
	if err != nil || parsed.Host != "github.com" {
		return "", "", 0, ErrInvalidUrl
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 4 || parts[2] != "pull" {
		return "", "", 0, ErrInvalidUrl
	}

	if !s.validRepos[parts[0]] {
		return "", "", 0, ErrInvalidRepo
	}

	number, err = strconv.Atoi(parts[3])
	if err != nil {
		return "", "", 0, ErrInvalidUrl
	}

	return parts[0], parts[1], number, nil
}

func (s *Server) approvePullRequest(r *http.Request, owner, repo string, number int) error {
	event := "APPROVE"
	_, _, err := s.client.PullRequests.CreateReview(r.Context(), owner, repo, number, &github.PullRequestReviewRequest{
		Event: &event,
	})
	return err
}

func (s *Server) mergePullRequest(r *http.Request, owner, repo string, number int) error {
	_, _, err := s.client.PullRequests.Merge(r.Context(), owner, repo, number, "", &github.PullRequestOptions{
		MergeMethod: "squash",
	})
	return err
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, response GitHubResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}
