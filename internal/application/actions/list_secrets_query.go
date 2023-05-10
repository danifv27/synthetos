package actions

import (
	"context"

	"fry.org/cmo/cli/internal/application/kms"
	"fry.org/cmo/cli/internal/application/logger"
	"github.com/speijnik/go-errortree"
)

// ListSecretsRequest query params
type ListSecretsRequest struct {
	SendCh  chan<- kms.Secret
	GroupID *string
}

type ListSecretsQuery interface {
	Handle(request ListSecretsRequest) error
}

// Implements ListSecretsQuery interface
type listSecretsQueryHandler struct {
	lgr   logger.Logger
	kmngr kms.KeyManager
}

// NewListSecretsQueryHandler Handler Constructor
func NewListSecretsQueryHandler(l logger.Logger, k kms.KeyManager) ListSecretsQuery {

	return listSecretsQueryHandler{
		lgr:   l,
		kmngr: k,
	}
}

func (h listSecretsQueryHandler) Handle(request ListSecretsRequest) error {
	var err, rcerror error
	var secrets []kms.Secret

	ctx := context.Background()
	if secrets, err = h.kmngr.ListSecrets(ctx, request.GroupID); err != nil {
		close(request.SendCh)
		return errortree.Add(rcerror, "Handle", err)
	}
	for _, s := range secrets {
		request.SendCh <- s
	}
	//Let's signal there is no more resources to process
	close(request.SendCh)

	return nil
}
