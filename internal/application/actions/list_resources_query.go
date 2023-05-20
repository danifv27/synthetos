package actions

import (
	"context"

	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/application/provider"
	"github.com/speijnik/go-errortree"
)

// ListResourcesRequest query params
type ListResourcesRequest struct {
	SendCh    chan<- provider.ResourceList
	Namespace *string
	Selector  string
	Concise   bool
}

type ListResourcesQuery interface {
	Handle(request ListResourcesRequest) error
}

// Implements ListResourcesQuery interface
type listResourcesQueryHandler struct {
	lgr   logger.Logger
	prvdr provider.ResourceProvider
}

// NewListResourcesQueryHandler Handler Constructor
func NewListResourcesQueryHandler(l logger.Logger, pr provider.ResourceProvider) ListResourcesQuery {

	return listResourcesQueryHandler{
		lgr:   l,
		prvdr: pr,
	}
}

func (h listResourcesQueryHandler) Handle(request ListResourcesRequest) error {
	var err, rcerror error

	ctx := context.Background()
	if err = h.prvdr.AllResources(ctx, request.SendCh, request.Namespace, request.Selector, request.Concise); err != nil {
		return errortree.Add(rcerror, "Handle", err)
	}

	return nil
}
