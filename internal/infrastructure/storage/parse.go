package storage

import (
	"fmt"
	"net/url"

	"fry.org/cmo/cli/internal/application/version"
	"fry.org/cmo/cli/internal/infrastructure/storage/embed"
	"github.com/speijnik/go-errortree"
)

func Parse(URI string) (version.Version, error) {
	var v version.Version
	var rcerror error

	u, err := url.Parse(URI)
	if err != nil {
		return nil, errortree.Add(rcerror, "parse", err)
	}
	if u.Scheme != "version" {
		return nil, errortree.Add(rcerror, "parse", fmt.Errorf("invalid scheme %s", URI))
	}

	switch u.Opaque {
	case "embed":
		v = embed.NewVersion()
	default:
		return nil, errortree.Add(rcerror, "parse", fmt.Errorf("unsupported version implementation %q", u.Opaque))
	}

	return v, nil
}
