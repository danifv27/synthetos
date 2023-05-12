package actions

import (
	"context"

	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/application/provider"
	"github.com/speijnik/go-errortree"
)

type ListManifestsObjectsRequest struct {
	SendCh    chan<- provider.Manifest
	InputPath string
}

type ListManifestsObjectsQuery interface {
	Handle(command ListManifestsObjectsRequest) error
}

type listManifestsObjectsQueryHandler struct {
	lgr   logger.Logger
	prvdr provider.ManifestProvider
}

// NewListManifestsObjectsQueryHandler Constructor
func NewListManifestsObjectsQueryHandler(l logger.Logger, pr provider.ManifestProvider) ListManifestsObjectsQuery {

	return listManifestsObjectsQueryHandler{
		lgr:   l,
		prvdr: pr,
	}
}

// Handle Handles the update request
func (h listManifestsObjectsQueryHandler) Handle(request ListManifestsObjectsRequest) error {
	var err, rcerror error

	ctx := context.Background()
	if err = h.prvdr.GetManifests(ctx, request.SendCh); err != nil {
		close(request.SendCh)
		return errortree.Add(rcerror, "Handle", err)
	}

	return nil
}
