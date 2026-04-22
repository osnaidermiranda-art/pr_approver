package service

import (
	"context"
	"fmt"
	"strings"

	"pr_approved/internal/ghclient"
)

const (
	Approve = "approve"
	Merge   = "merge"
	Both    = "both"
)

var validActions = map[string]bool{
	Approve: true,
	Merge:   true,
	Both:    true,
}

type Service interface {
	IsValidOwner(owner string) bool
	IsValidAction(action string) bool
	HandlePR(ctx context.Context, owner, repo string, number int, action string) error
}

type GitHubService struct {
	clients     map[string]ghclient.Client
	validOwners map[string]bool
}

func NewGitHubService(clients map[string]ghclient.Client, validOwners map[string]bool) *GitHubService {
	return &GitHubService{clients: clients, validOwners: validOwners}
}

func (s *GitHubService) IsValidOwner(owner string) bool {
	return s.validOwners[strings.ToLower(owner)]
}

func (s *GitHubService) IsValidAction(action string) bool {
	return validActions[action]
}

func (s *GitHubService) HandlePR(ctx context.Context, owner, repo string, number int, action string) error {
	client, err := s.clientFor(owner)
	if err != nil {
		return err
	}

	if action == Approve || action == Both {
		if err := client.ApprovePR(ctx, owner, repo, number); err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToApprovePr, err)
		}
	}

	if action == Merge || action == Both {
		if err := client.MergePR(ctx, owner, repo, number); err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToMergePr, err)
		}
	}

	return nil
}

func (s *GitHubService) clientFor(owner string) (ghclient.Client, error) {
	c, ok := s.clients[strings.ToLower(owner)]
	if !ok {
		return nil, ErrNoTokenForOwner
	}
	return c, nil
}
