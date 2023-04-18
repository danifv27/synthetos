package kms

import (
	"errors"
	"fmt"
	"net/url"

	"fry.org/cmo/cli/internal/application/kms"
	"fry.org/cmo/cli/internal/application/logger"
	"github.com/speijnik/go-errortree"
) // A KmsOption applies optional changes to the kms implementation
type KmsOption interface {
	Apply(t interface{}) error
}

// KmsOptionFn is function that adheres to the KmsOption interface.
type KmsOptionFn func(t interface{}) error

func (o KmsOptionFn) Apply(t interface{}) error {

	return o(t)
}

// Parse the uri string and returns the proper kms.KeyManager implementation
// Available uris:
// keymanager:fortanix?endpoint=<endpoint>&apikey=<apikey>
func Parse(URI string, l logger.Logger) (kms.KeyManager, error) {
	var k kms.KeyManager
	var err, rcerror error
	var u *url.URL

	u, err = url.Parse(URI)
	if err != nil {
		rcerror = errortree.Add(rcerror, "kms.Parse", err)
		return nil, rcerror
	}
	if u.Scheme != "keymanager" {
		rcerror = errortree.Add(rcerror, "kms.Parse", fmt.Errorf("invalid scheme %s", URI))
		return nil, rcerror
	}
	switch u.Opaque {
	case "fortanix":
		options := []KmsOption{
			WithLogger(l),
		}
		apikey := u.Query().Get("apikey")
		if apikey == "" {
			rcerror = errortree.Add(rcerror, "kms.Parse", fmt.Errorf("invalid scheme %s", URI))
			return nil, rcerror
		}
		options = append(options,
			WithApikey(apikey),
		)
		endpoint := u.Query().Get("endpoint")
		if endpoint == "" {
			l.WithFields(logger.Fields{
				"endpoint": endpoint,
			}).Debug("Usind default kms endpoint")
			endpoint = "https://kms.adidas.com"
		}
		options = append(options,
			WithEndpoint(endpoint),
		)
		if k, err = NewFortanixKms(options...); err != nil {
			rcerror = errortree.Add(rcerror, "kms.Parse", err)
			return nil, rcerror
		}
	default:
		rcerror = errortree.Add(rcerror, "kms.Parse", fmt.Errorf("unsupported customizer implementation %q", u.Opaque))
		return nil, rcerror
	}

	return k, nil
}

func WithLogger(l logger.Logger) KmsOption {

	return KmsOptionFn(func(i interface{}) error {
		var rcerror error
		var f *fortanixClient
		var ok bool

		if f, ok = i.(*fortanixClient); ok {
			f.l = l
			return nil
		}

		return errortree.Add(rcerror, "fortanix.WithLogger", errors.New("type mismatch, fortanixClient expected"))
	})
}
