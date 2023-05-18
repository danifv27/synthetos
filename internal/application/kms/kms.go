package kms

import "context"

type KeyManager interface {
	// Authenticate(ctx context.Context) error
	// Get(ctx context.Context) error
	ListGroups(ctx context.Context) ([]Group, error)
	ListSecrets(ctx context.Context, groupID *string) ([]Secret, error)
	DecryptSecret(ctx context.Context, id *string) (Secret, error)
}

type Group struct {
	CreatedAt   string  `json:"created_at"`
	Description *string `json:"description,omitempty"`
	GroupID     string  `json:"group_id"`
	Name        string  `json:"name"`
}

type Secret struct {
	GroupID     *string  `json:"group_id,omitempty"`
	SecretID    *string  `json:"kid,omitempty"`
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	CreatedAt   string   `json:"created_at"`
	LastusedAt  string   `json:"lastused_at"`
	Splitted    []string `json:"splitted,omitempty"` //Blob splitted line by line
	Value       string   `json:"value,omitempty"`    //Blob as a single string
	Blob        *[]byte  `json:"-"`
}
