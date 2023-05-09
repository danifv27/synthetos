package kms

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"time"

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

	//Remove TLS certs validation
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	clh := &http.Client{Transport: tr}

	f := fortanixClient{
		client: sdkms.Client{
			HTTPClient: clh,
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
			// f.client.Auth = sdkms.APIKey(apikey)
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

func (f *fortanixClient) ListGroups(ctx context.Context) ([]kms.Group, error) {
	var rcerror error
	var groups []kms.Group

	cx, cancelfn := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelfn()
	// Establish a session
	if _, err := f.client.AuthenticateWithAPIKey(cx, f.apikey); err != nil {
		f.l.WithFields(logger.Fields{
			"err": err,
		}).Info("Authentication failed")
		return []kms.Group{}, errortree.Add(rcerror, "fortanix.ListGroups", err)
	}
	// Terminate the session on exit
	defer f.client.TerminateSession(ctx)
	// List groups
	gs, err := f.client.ListGroups(ctx)
	if err != nil {
		return []kms.Group{}, errortree.Add(rcerror, "fortanix.ListGroups", err)
	}
	for _, g := range gs {
		groups = append(groups, kms.Group{
			CreatedAt:   string(g.CreatedAt),
			Description: g.Description,
			Name:        g.Name,
			GroupID:     g.GroupID,
		})

	}

	return groups, nil
}

func (f *fortanixClient) ListSecrets(ctx context.Context) ([]kms.Secret, error) {
	var rcerror error
	var secrets []kms.Secret

	cx, cancelfn := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelfn()
	// Establish a session
	if _, err := f.client.AuthenticateWithAPIKey(cx, f.apikey); err != nil {
		f.l.WithFields(logger.Fields{
			"err": err,
		}).Info("Authentication failed")
		return []kms.Secret{}, errortree.Add(rcerror, "fortanix.ListSecrets", err)
	}
	// Terminate the session on exit
	defer f.client.TerminateSession(ctx)
	// List groups
	gs, err := f.client.ListSobjects(ctx, nil)
	if err != nil {
		return []kms.Secret{}, errortree.Add(rcerror, "fortanix.ListSecrets", err)
	}
	for _, g := range gs {
		secrets = append(secrets, kms.Secret{
			CreatedAt:   string(g.CreatedAt),
			LastusedAt:  string(g.LastusedAt),
			Description: g.Description,
			Name:        g.Name,
			GroupID:     g.GroupID,
			Blob:        g.Value,
			SecretID:    g.Kid,
		})
	}

	return secrets, nil
}

// func sobjectToString(sobject *sdkms.Sobject) string {
// 	created, err := sobject.CreatedAt.Std()
// 	if err != nil {
// 		return err.Error()
// 	}
// 	return fmt.Sprintf("{ %v %#v group(%v) enabled: %v created: %v }",
// 		*sobject.Kid, *sobject.Name, *sobject.GroupID, sobject.Enabled,
// 		created.Local())
// }

func (f *fortanixClient) Decrypt(ctx context.Context) error {
	var rcerror error

	return errortree.Add(rcerror, "fortanix.Decrypt", errors.New("method not implemented"))
}
