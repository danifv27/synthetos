package actions

import (
	"fry.org/cmo/cli/internal/application/kms"
)

// ListGroupsRequest query params
type ListGroupsRequest struct {
}

type ListGroupsResult struct {
	Groups []kms.Group
}

type ListGroupsHandler interface {
	Handle(request ListGroupsRequest) (ListGroupsResult, error)
}

// Implements ListGroupsHandler interface
type listGroupsHandler struct {
}

// NewListGroupsHandler Handler Constructor
func NewListGroupsHandler() ListGroupsHandler {

	return listGroupsHandler{}
}

func (h listGroupsHandler) Handle(request ListGroupsRequest) (ListGroupsResult, error) {

	return ListGroupsResult{}, nil
}
