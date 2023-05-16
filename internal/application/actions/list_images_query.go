package actions

import (
	"context"
	"fmt"

	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/application/provider"
	"github.com/speijnik/go-errortree"
)

// ListImagesRequest query params
type ListImagesRequest struct {
	SendCh chan<- provider.Image
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
	var images []provider.Image

	ctx := context.Background()
	if images, err = h.prvdr.AllImages(ctx); err != nil {
		close(request.SendCh)
		return errortree.Add(rcerror, "Handle", err)
	}
	fmt.Printf("[DBG]images: %v", images)
	//Let's signal there is no more resources to process
	close(request.SendCh)

	return nil
}
