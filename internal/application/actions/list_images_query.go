package actions

import (
	"context"

	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/application/provider"
	"github.com/speijnik/go-errortree"
)

// ListImagesRequest query params
type ListImagesRequest struct {
	SendCh   chan<- provider.Image
	Selector string
}

type ListImagesQuery interface {
	Handle(request ListImagesRequest) error
}

// Implements ListImagesQuery interface
type listImagesQueryHandler struct {
	lgr   logger.Logger
	prvdr provider.ResourceProvider
}

// NewListImagesQueryHandler Handler Constructor
func NewListImagesQueryHandler(l logger.Logger, pr provider.ResourceProvider) ListImagesQuery {

	return listImagesQueryHandler{
		lgr:   l,
		prvdr: pr,
	}
}

func (h listImagesQueryHandler) Handle(request ListImagesRequest) error {
	var err, rcerror error

	ctx := context.Background()
	if err = h.prvdr.AllImages(ctx, request.SendCh, request.Selector); err != nil {
		return errortree.Add(rcerror, "Handle", err)
	}

	return nil
}
