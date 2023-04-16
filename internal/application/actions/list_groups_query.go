package actions

import (
	"context"
	"errors"

	"fry.org/cmo/cli/internal/application/kms"
	"fry.org/cmo/cli/internal/application/logger"
	"github.com/speijnik/go-errortree"
)

// ListGroupsRequest query params
type ListGroupsRequest struct {
}

type ListGroupsResult struct {
	Groups []kms.Group
}

type ListGroupsQueryHandler interface {
	Handle(request ListGroupsRequest) (ListGroupsResult, error)
}

// Implements ListGroupsHandler interface
type listGroupsQueryHandler struct {
	l logger.Logger
	k kms.KeyManager
}

// NewListGroupsQueryHandler Handler Constructor
func NewListGroupsQueryHandler(lgr logger.Logger, k kms.KeyManager) ListGroupsQueryHandler {

	return listGroupsQueryHandler{
		l: lgr,
	}
}

func (h listGroupsQueryHandler) Handle(request ListGroupsRequest) (ListGroupsResult, error) {
	var rcerror error

	ctx := context.Background()
	if err := h.k.List(ctx); err != nil {
		return ListGroupsResult{}, errortree.Add(rcerror, "Handle", err)
	}

	return ListGroupsResult{}, errortree.Add(rcerror, "Handle", errors.New("method not implemented"))
}
