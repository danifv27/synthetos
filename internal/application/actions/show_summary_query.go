package actions

import (
	"context"
	"errors"

	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/application/provider"
	"github.com/speijnik/go-errortree"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ShowSummaryRequest query params
type ShowSummaryRequest struct {
	Location string
	Selector string
}

type ShowSummaryResult struct {
	Items []provider.Summary
}

type ShowSummaryQueryHandler interface {
	Handle(request ShowSummaryRequest) (ShowSummaryResult, error)
}

// Implements ShowSummaryHandler interface
type showSummaryQueryHandler struct {
	lgr      logger.Logger
	provider provider.ResourceProvider
}

// NewShowSummaryQueryHandler Handler Constructor
func NewShowSummaryQueryHandler(l logger.Logger, pr provider.ResourceProvider) ShowSummaryQueryHandler {

	return showSummaryQueryHandler{
		lgr:      l,
		provider: pr,
	}
}

func summarize(u *unstructured.Unstructured) (provider.Summary, error) {
	var rcerror error

	s := provider.Summary{
		APIVersion: u.GetAPIVersion(),
		Kind:       u.GetKind(),
	}

	nameRaw := u.GetName()
	if nameRaw != "" {
		s.Name = nameRaw
		return s, nil
	}

	generateNameRaw := u.GetGenerateName()
	if generateNameRaw != "" {
		s.Name = generateNameRaw
		return s, nil
	}

	return provider.Summary{}, errortree.Add(rcerror, "summarize", errors.New("unable to find object name"))
}

func (h showSummaryQueryHandler) Handle(request ShowSummaryRequest) (ShowSummaryResult, error) {
	var err, rcerror error
	var resources []*unstructured.Unstructured

	ctx := context.Background()
	rc := ShowSummaryResult{
		Items: make([]provider.Summary, 0),
	}
	if resources, err = h.provider.GetResources(ctx, request.Location, request.Selector); err != nil {
		return ShowSummaryResult{}, errortree.Add(rcerror, "Handle", err)
	}
	for _, r := range resources {
		if s, err := summarize(r); err != nil {
			return ShowSummaryResult{}, errortree.Add(rcerror, "Handle", err)
		} else {
			rc.Items = append(rc.Items, s)
		}
	}

	return rc, nil
}
