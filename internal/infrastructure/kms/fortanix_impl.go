package kms

import (
	"context"
	"errors"
	"net/http"

	"fry.org/cmo/cli/internal/application/kms"
	"fry.org/cmo/cli/internal/application/logger"
	"github.com/fortanix/sdkms-client-go/sdkms"
	"github.com/speijnik/go-errortree"
)

type fortanixClient struct {
	l      logger.Logger
	apikey string
	client sdkms.Client
}

// NewFortanixKms creates a new CucumberExporter
func NewFortanixKms(opts ...KmsOption) (kms.KeyManager, error) {
	var rcerror error

	f := fortanixClient{
		client: sdkms.Client{
			HTTPClient: http.DefaultClient,
		},
	}
	// Loop through each option
	for _, option := range opts {
		if err := option.Apply(&f); err != nil {
			return nil, errortree.Add(rcerror, "NewFortanixKms", err)
		}
	}

	return &f, nil
}

func WithApikey(apikey string) KmsOption {

	return KmsOptionFn(func(i interface{}) error {
		var rcerror error
		var f *fortanixClient
		var ok bool

		if f, ok = i.(*fortanixClient); ok {
			f.apikey = apikey
			f.client.Auth = sdkms.APIKey(apikey)
			return nil
		}

		return errortree.Add(rcerror, "fortanix.WithApikey", errors.New("type mismatch, fortanixClient expected"))
	})
}

func WithEndpoint(url string) KmsOption {

	return KmsOptionFn(func(i interface{}) error {
		var rcerror error
		var f *fortanixClient
		var ok bool

		if f, ok = i.(*fortanixClient); ok {
			f.client.Endpoint = url
			return nil
		}

		return errortree.Add(rcerror, "fortanix.WithEndpoint", errors.New("type mismatch, fortanixClient expected"))
	})
}

// TODO: implement keymanager interface
func (f *fortanixClient) Get(ctx context.Context) error {
	var rcerror error

	return errortree.Add(rcerror, "fortanix.Get", errors.New("method not implemented"))
}

func (f *fortanixClient) List(ctx context.Context) error {
	var rcerror error

	// Establish a session
	_, err := f.client.AuthenticateWithAPIKey(ctx, f.apikey)
	if err != nil {
		f.l.WithFields(logger.Fields{
			"err": err,
		}).Info("Authentication failed")
		return errortree.Add(rcerror, "fortanix.List", err)
	}
	// Terminate the session on exit
	defer f.client.TerminateSession(ctx)

	return errortree.Add(rcerror, "fortanix.List", errors.New("method not implemented"))
}

func (f *fortanixClient) Decrypt(ctx context.Context) error {
	var rcerror error

	return errortree.Add(rcerror, "fortanix.Decrypt", errors.New("method not implemented"))
}
