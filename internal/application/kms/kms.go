package kms

import "context"

type KeyManager interface {
	// Authenticate(ctx context.Context) error
	Get(ctx context.Context) error
	ListGroups(ctx context.Context) ([]Group, error)
	Decrypt(ctx context.Context) error
}

type Group struct {
	CreatedAt   string  `json:"created_at"`
	Description *string `json:"description,omitempty"`
	GroupID     string  `json:"group_id"`
	Name        string  `json:"name"`
}
