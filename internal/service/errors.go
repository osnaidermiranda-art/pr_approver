package service

import "errors"

var (
	ErrInvalidUrl        = errors.New("Invalid URL")
	ErrInvalidRepo       = errors.New("Invalid repository")
	ErrInvalidAction     = errors.New("Invalid action")
	ErrNoTokenForOwner   = errors.New("No GitHub token configured for this owner")
	ErrFailedToApprovePr = errors.New("Failed to approve pull request")
	ErrFailedToMergePr   = errors.New("Failed to merge pull request")
)
