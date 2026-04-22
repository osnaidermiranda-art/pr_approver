package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"pr_approved/internal/handler"
	"pr_approved/internal/service"
)

type mockService struct {
	isValidOwnerFn func(owner string) bool
	isValidActionFn func(action string) bool
	handlePRFn     func(ctx context.Context, owner, repo string, number int, action string) error
}

func (m *mockService) IsValidOwner(owner string) bool  { return m.isValidOwnerFn(owner) }
func (m *mockService) IsValidAction(action string) bool { return m.isValidActionFn(action) }
func (m *mockService) HandlePR(ctx context.Context, owner, repo string, number int, action string) error {
	return m.handlePRFn(ctx, owner, repo, number, action)
}

func okService() *mockService {
	return &mockService{
		isValidOwnerFn:  func(_ string) bool { return true },
		isValidActionFn: func(_ string) bool { return true },
		handlePRFn:      func(_ context.Context, _, _ string, _ int, _ string) error { return nil },
	}
}

func post(h *handler.GitHubHandler, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/git-hub", bytes.NewReader(b))
	w := httptest.NewRecorder()
	h.HandlePullRequest(w, req)
	return w
}

func TestHandlePullRequest_InvalidURL(t *testing.T) {
	svc := okService()
	h := handler.NewGitHubHandler(svc)

	for _, url := range []string{
		"not-a-url",
		"https://notgithub.com/org/repo/pull/1",
		"https://github.com/org/repo/issues/1",
		"https://github.com/org/repo/pull/abc",
	} {
		w := post(h, map[string]string{"url": url, "action": "both"})
		if w.Code != http.StatusBadRequest {
			t.Errorf("url %q: expected 400, got %d", url, w.Code)
		}
	}
}

func TestHandlePullRequest_InvalidRepo(t *testing.T) {
	svc := okService()
	svc.isValidOwnerFn = func(_ string) bool { return false }
	h := handler.NewGitHubHandler(svc)

	w := post(h, map[string]string{"url": "https://github.com/unknown-org/repo/pull/1", "action": "both"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	assertErrorMessage(t, w, service.ErrInvalidRepo.Error())
}

func TestHandlePullRequest_InvalidAction(t *testing.T) {
	svc := okService()
	svc.isValidActionFn = func(_ string) bool { return false }
	h := handler.NewGitHubHandler(svc)

	w := post(h, map[string]string{"url": "https://github.com/myorg/repo/pull/1", "action": "invalid"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	assertErrorMessage(t, w, service.ErrInvalidAction.Error())
}

func TestHandlePullRequest_DefaultsActionToBoth(t *testing.T) {
	calledWith := ""
	svc := okService()
	svc.handlePRFn = func(_ context.Context, _, _ string, _ int, action string) error {
		calledWith = action
		return nil
	}
	h := handler.NewGitHubHandler(svc)

	w := post(h, map[string]string{"url": "https://github.com/myorg/repo/pull/1"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if calledWith != service.Both {
		t.Fatalf("expected action %q, got %q", service.Both, calledWith)
	}
}

func TestHandlePullRequest_Success(t *testing.T) {
	cases := []struct {
		action  string
		message string
	}{
		{service.Approve, "Pull request approved"},
		{service.Merge, "Pull request merged"},
		{service.Both, "Pull request approved and merged"},
	}

	for _, tc := range cases {
		h := handler.NewGitHubHandler(okService())
		w := post(h, map[string]string{"url": "https://github.com/myorg/repo/pull/42", "action": tc.action})

		if w.Code != http.StatusOK {
			t.Errorf("action %q: expected 200, got %d", tc.action, w.Code)
		}
		assertMessage(t, w, tc.message)
	}
}

func TestHandlePullRequest_ServiceError(t *testing.T) {
	cases := []struct {
		err            error
		expectedStatus int
	}{
		{service.ErrNoTokenForOwner, http.StatusBadRequest},
		{service.ErrFailedToApprovePr, http.StatusInternalServerError},
		{service.ErrFailedToMergePr, http.StatusInternalServerError},
	}

	for _, tc := range cases {
		svc := okService()
		svc.handlePRFn = func(_ context.Context, _, _ string, _ int, _ string) error { return tc.err }
		h := handler.NewGitHubHandler(svc)

		w := post(h, map[string]string{"url": "https://github.com/myorg/repo/pull/1", "action": "both"})
		if w.Code != tc.expectedStatus {
			t.Errorf("err %v: expected %d, got %d", tc.err, tc.expectedStatus, w.Code)
		}
		assertErrorMessage(t, w, tc.err.Error())
	}
}

func decodeResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]string {
	t.Helper()
	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return resp
}

func assertMessage(t *testing.T, w *httptest.ResponseRecorder, expected string) {
	t.Helper()
	resp := decodeResponse(t, w)
	if resp["message"] != expected {
		t.Errorf("expected message %q, got %q", expected, resp["message"])
	}
}

func assertErrorMessage(t *testing.T, w *httptest.ResponseRecorder, expected string) {
	t.Helper()
	resp := decodeResponse(t, w)
	if resp["message"] != expected {
		t.Errorf("expected message %q, got %q", expected, resp["message"])
	}
	if resp["status"] != "error" {
		t.Errorf("expected status=error, got %q", resp["status"])
	}
}

// ensure *GitHubService satisfies Service at compile time
var _ service.Service = (*mockService)(nil)

// ensure errors are comparable with errors.Is
var _ = errors.Is(service.ErrFailedToApprovePr, service.ErrFailedToApprovePr)
