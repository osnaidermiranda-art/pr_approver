package ghclient

import (
	"context"

	gogithub "github.com/google/go-github/v84/github"
)

type Client interface {
	ApprovePR(ctx context.Context, owner, repo string, number int) error
	MergePR(ctx context.Context, owner, repo string, number int) error
}

type client struct {
	gh *gogithub.Client
}

func New(token string) Client {
	return &client{gh: gogithub.NewClient(nil).WithAuthToken(token)}
}

func (c *client) ApprovePR(ctx context.Context, owner, repo string, number int) error {
	event := "APPROVE"
	_, _, err := c.gh.PullRequests.CreateReview(ctx, owner, repo, number, &gogithub.PullRequestReviewRequest{
		Event: &event,
	})
	return err
}

func (c *client) MergePR(ctx context.Context, owner, repo string, number int) error {
	_, _, err := c.gh.PullRequests.Merge(ctx, owner, repo, number, "", &gogithub.PullRequestOptions{
		MergeMethod: "squash",
	})
	return err
}
