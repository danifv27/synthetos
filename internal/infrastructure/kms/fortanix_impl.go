package kms

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"fry.org/cmo/cli/internal/application/kms"
	"fry.org/cmo/cli/internal/application/logger"
	"github.com/fortanix/sdkms-client-go/sdkms"
	"github.com/google/uuid"
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

// // TODO: implement keymanager interface
// func (f *fortanixClient) Get(ctx context.Context) error {
// 	var rcerror error

// 	return errortree.Add(rcerror, "fortanix.Get", errors.New("method not implemented"))
// }

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

func (f *fortanixClient) listSobjects(ctx context.Context, groupID *string) ([]sdkms.Sobject, error) {
	var rcerror, err error
	var objects []sdkms.Sobject

	cx, cancelfn := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelfn()
	// Establish a session
	if _, err = f.client.AuthenticateWithAPIKey(cx, f.apikey); err != nil {
		f.l.WithFields(logger.Fields{
			"err": err,
		}).Info("Authentication failed")
		return objects, errortree.Add(rcerror, "fortanix.listSobjects", err)
	}
	// Terminate the session on exit
	defer f.client.TerminateSession(ctx)
	// List groups
	queryParams := sdkms.ListSobjectsParams{
		Sort: sdkms.SobjectSort{
			ByName: &sdkms.SobjectSortByName{},
		},
	}
	if groupID != nil {
		queryParams.GroupID = groupID
	}
	if objects, err = f.client.ListSobjects(ctx, &queryParams); err != nil {
		return objects, errortree.Add(rcerror, "fortanix.listSobjects", err)
	}

	return objects, nil
}

func (f *fortanixClient) ListSecrets(ctx context.Context, groupID *string) ([]kms.Secret, error) {
	var rcerror error
	var secrets []kms.Secret

	gs, err := f.listSobjects(ctx, groupID)
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
			// Blob:        g.Value,
			SecretID: g.Kid,
		})
	}

	return secrets, nil
}

// FIXME: If there a secret with same name in two different groups, we are going into trouble
func (f *fortanixClient) DecryptSecret(ctx context.Context, id *string) (kms.Secret, error) {
	var rcerror, err error
	var gs []sdkms.Sobject

	if gs, err = f.listSobjects(ctx, nil); err != nil {
		return kms.Secret{}, errortree.Add(rcerror, "fortanix.DecryptSecret", err)
	}
	// Parse the string as a UUID
	_, err = uuid.Parse(*id)
	for _, g := range gs {
		if ((err != nil) && (*g.Name == *id)) || (*g.Kid == *id) {
			return kms.Secret{
				CreatedAt:   string(g.CreatedAt),
				LastusedAt:  string(g.LastusedAt),
				Description: g.Description,
				Name:        g.Name,
				GroupID:     g.GroupID,
				Blob:        g.Value,
				SecretID:    g.Kid,
			}, nil
		}
	}

	return kms.Secret{}, errortree.Add(rcerror, "fortanix.DecryptSecret", fmt.Errorf("secret %s not found", *id))
}
