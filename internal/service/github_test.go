package service_test

import (
	"context"
	"errors"
	"testing"

	"pr_approved/internal/ghclient"
	"pr_approved/internal/service"
)

type mockClient struct {
	approveFn func(ctx context.Context, owner, repo string, number int) error
	mergeFn   func(ctx context.Context, owner, repo string, number int) error
}

func (m *mockClient) ApprovePR(ctx context.Context, owner, repo string, number int) error {
	return m.approveFn(ctx, owner, repo, number)
}

func (m *mockClient) MergePR(ctx context.Context, owner, repo string, number int) error {
	return m.mergeFn(ctx, owner, repo, number)
}

func okClient() ghclient.Client {
	return &mockClient{
		approveFn: func(_ context.Context, _, _ string, _ int) error { return nil },
		mergeFn:   func(_ context.Context, _, _ string, _ int) error { return nil },
	}
}

func newSvc(owner string, client ghclient.Client) *service.GitHubService {
	return service.NewGitHubService(
		map[string]ghclient.Client{owner: client},
		map[string]bool{owner: true},
	)
}

func TestHandlePR_Approve(t *testing.T) {
	approved := false
	client := &mockClient{
		approveFn: func(_ context.Context, _, _ string, _ int) error { approved = true; return nil },
		mergeFn:   func(_ context.Context, _, _ string, _ int) error { t.Fatal("merge should not be called"); return nil },
	}
	svc := newSvc("myorg", client)

	if err := svc.HandlePR(context.Background(), "myorg", "repo", 1, service.Approve); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !approved {
		t.Fatal("expected ApprovePR to be called")
	}
}

func TestHandlePR_Merge(t *testing.T) {
	merged := false
	client := &mockClient{
		approveFn: func(_ context.Context, _, _ string, _ int) error { t.Fatal("approve should not be called"); return nil },
		mergeFn:   func(_ context.Context, _, _ string, _ int) error { merged = true; return nil },
	}
	svc := newSvc("myorg", client)

	if err := svc.HandlePR(context.Background(), "myorg", "repo", 1, service.Merge); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !merged {
		t.Fatal("expected MergePR to be called")
	}
}

func TestHandlePR_Both(t *testing.T) {
	approved, merged := false, false
	client := &mockClient{
		approveFn: func(_ context.Context, _, _ string, _ int) error { approved = true; return nil },
		mergeFn:   func(_ context.Context, _, _ string, _ int) error { merged = true; return nil },
	}
	svc := newSvc("myorg", client)

	if err := svc.HandlePR(context.Background(), "myorg", "repo", 1, service.Both); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !approved || !merged {
		t.Fatalf("expected both approve and merge, got approve=%v merge=%v", approved, merged)
	}
}

func TestHandlePR_ApproveError(t *testing.T) {
	client := &mockClient{
		approveFn: func(_ context.Context, _, _ string, _ int) error { return errors.New("api error") },
		mergeFn:   func(_ context.Context, _, _ string, _ int) error { t.Fatal("merge should not be called"); return nil },
	}
	svc := newSvc("myorg", client)

	err := svc.HandlePR(context.Background(), "myorg", "repo", 1, service.Both)
	if !errors.Is(err, service.ErrFailedToApprovePr) {
		t.Fatalf("expected ErrFailedToApprovePr, got %v", err)
	}
}

func TestHandlePR_MergeError(t *testing.T) {
	client := &mockClient{
		approveFn: func(_ context.Context, _, _ string, _ int) error { return nil },
		mergeFn:   func(_ context.Context, _, _ string, _ int) error { return errors.New("api error") },
	}
	svc := newSvc("myorg", client)

	err := svc.HandlePR(context.Background(), "myorg", "repo", 1, service.Both)
	if !errors.Is(err, service.ErrFailedToMergePr) {
		t.Fatalf("expected ErrFailedToMergePr, got %v", err)
	}
}

func TestHandlePR_NoTokenForOwner(t *testing.T) {
	svc := newSvc("myorg", okClient())

	err := svc.HandlePR(context.Background(), "other-org", "repo", 1, service.Approve)
	if !errors.Is(err, service.ErrNoTokenForOwner) {
		t.Fatalf("expected ErrNoTokenForOwner, got %v", err)
	}
}

func TestIsValidOwner(t *testing.T) {
	svc := newSvc("myorg", okClient())

	tests := []struct {
		owner string
		valid bool
	}{
		{"myorg", true},
		{"MYORG", true},
		{"MyOrg", true},
		{"unknown", false},
	}

	for _, tt := range tests {
		if svc.IsValidOwner(tt.owner) != tt.valid {
			t.Errorf("IsValidOwner(%q) = %v, want %v", tt.owner, !tt.valid, tt.valid)
		}
	}
}

func TestIsValidAction(t *testing.T) {
	svc := newSvc("myorg", okClient())

	for _, valid := range []string{service.Approve, service.Merge, service.Both} {
		if !svc.IsValidAction(valid) {
			t.Fatalf("expected %q to be valid", valid)
		}
	}
	if svc.IsValidAction("invalid") {
		t.Fatal("expected 'invalid' to be invalid action")
	}
}
