package kms

import "context"

type KeyManager interface {
	Authenticate(ctx context.Context) error
	Get(ctx context.Context) error
	List(ctx context.Context) error
	Decrypt(ctx context.Context) error
}
