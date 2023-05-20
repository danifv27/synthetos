package actions

import (
	"errors"
	"fmt"
	"path/filepath"

	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/application/provider"
	"github.com/speijnik/go-errortree"
)

// FilterResourceRequest query params
type FilterResourceRequest struct {
	InputCh  <-chan provider.Summary
	OutputCh chan<- provider.Summary
}

type FilterResourceCommand interface {
	Handle(request FilterResourceRequest) error
	WithIncludeMatch(pattern string) error
	WithExcludeMatch(pattern string) error
}

// Implements FilterResourceCommand interface
type filterResourceCommandHandler struct {
	lgr          logger.Logger
	matchInclude []string //follows the kubectl api-resource output apiversion.kind
	matchExclude []string //follows the kubectl api-resource output apiversion.kind
}

// NewFilterResourceCommandHandler Handler Constructor
func NewFilterResourceCommandHandler(l logger.Logger) FilterResourceCommand {

	return &filterResourceCommandHandler{
		lgr: l,
	}
}

// pattern follows the kubectl api-resource output apiversion.kind
func (h *filterResourceCommandHandler) WithIncludeMatch(pattern string) error {

	h.matchInclude = append(h.matchInclude, pattern)

	return nil
}

// pattern follows the kubectl api-resource output apiversion.kind
func (h *filterResourceCommandHandler) WithExcludeMatch(pattern string) error {

	h.matchExclude = append(h.matchExclude, pattern)

	return nil
}

// matchPattern checks whether a string matches a pattern.
func matchPattern(pattern string, s string) (bool, error) {
	var err, rcerror error
	var matched bool

	if pattern == "*" {
		return true, nil
	}
	if matched, err = filepath.Match(pattern, s); err != nil {
		rcerror = errortree.Add(rcerror, "matchPattern", err)
		return false, rcerror
	}

	return matched, nil
}

// allowed ckeck the key againsts include and exclude pattern.
func (h *filterResourceCommandHandler) allowed(key string) (bool, error) {
	var err, rcerror error
	var matched bool

	if (len(h.matchExclude) == 0) && (len(h.matchInclude) == 0) {
		// If there is no filtering patterns we allow all
		return true, nil
	}
	if len(h.matchExclude) > 0 {
		for _, element := range h.matchExclude {
			if matched, err = matchPattern(element, key); err != nil {
				rcerror = errortree.Add(rcerror, "allowed", err)
				return true, rcerror
			} else if matched {
				// Excludes has precedence over include
				return false, nil
			}
		}
	}
	if len(h.matchInclude) > 0 {
		for _, element := range h.matchInclude {
			if matched, err = matchPattern(element, key); err != nil {
				rcerror = errortree.Add(rcerror, "allowed", err)
				return false, rcerror
			}
			if matched {
				// It's not excluded, and is included, break the loop
				return matched, nil
			}
		}
	} else {
		// There is no include filter and the value is not excluded so it's allowed
		return true, nil
	}
	// It's not excluded, nor included
	rcerror = errortree.Add(rcerror, "allowed", fmt.Errorf("allowed: %s is neither included nor excluded", key))

	return false, rcerror
}

func (h *filterResourceCommandHandler) Handle(request FilterResourceRequest) error {
	var rcerror error

	return errortree.Add(rcerror, "Handle", errors.New("Handle method not implemented"))
}
