package actions

import (
	"context"

	"fry.org/cmo/cli/internal/application/kms"
	"fry.org/cmo/cli/internal/application/logger"
	"github.com/speijnik/go-errortree"
)

// DecryptSecretRequest query params
type DecryptSecretRequest struct {
	ReceiveCh <-chan kms.Secret
	SendCh    chan<- kms.Secret
	SecretID  *string
	Name      *string
}

type DecryptSecretQuery interface {
	Handle(request DecryptSecretRequest) error
}

// Implements DecryptSecretsQuery interface
type decryptSecretQueryHandler struct {
	lgr   logger.Logger
	kmngr kms.KeyManager
}

// NewDecryptSecretQueryHandler Handler Constructor
func NewDecryptSecretQueryHandler(l logger.Logger, k kms.KeyManager) DecryptSecretQuery {

	return decryptSecretQueryHandler{
		lgr:   l,
		kmngr: k,
	}
}

func (h decryptSecretQueryHandler) Handle(request DecryptSecretRequest) error {
	var err, rcerror error
	var secret kms.Secret

	ctx := context.Background()
	if request.Name != nil {
		if secret, err = h.kmngr.DecryptSecret(ctx, request.Name); err != nil {
			close(request.SendCh)
			return errortree.Add(rcerror, "Handle", err)
		}
	}
	request.SendCh <- secret
	//Let's signal there is no more resources to process
	close(request.SendCh)

	return nil
}
