package actions

import (
	"context"
	b64 "encoding/base64"
	"errors"
	"strconv"
	"strings"
	"unicode"

	"fry.org/cmo/cli/internal/application/kms"
	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/application/provider"
	"github.com/speijnik/go-errortree"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/kubectl/pkg/scheme"
)

type DecryptManifestsRequest struct {
	ReceiverCh <-chan provider.Manifest
	SendCh     chan<- provider.Manifest
}

type DecryptManifestsCommand interface {
	Handle(command DecryptManifestsRequest) error
}

type decryptManifestsCommandHandler struct {
	lgr   logger.Logger
	kmngr kms.KeyManager
}

// NewDecryptManifestsCommandHandler Constructor
func NewDecryptManifestsCommandHandler(l logger.Logger, k kms.KeyManager) DecryptManifestsCommand {

	return decryptManifestsCommandHandler{
		lgr:   l,
		kmngr: k,
	}
}

func setSecretDataFromKeyValueString(s *v1.Secret, data string) error {
	var err error
	var decoded []byte
	var foundEncodingSchema bool
	var encodingSchema string
	var unquoted string

	lastQuote := rune(0)
	f := func(c rune) bool {
		switch {
		case c == lastQuote:
			lastQuote = rune(0)
			return false
		case lastQuote != rune(0):
			return false
		case unicode.In(c, unicode.Quotation_Mark):
			lastQuote = c
			return false
		case c == 0x000A, c == 0x000B, c == 0x000C, c == 0x000D, c == 0x0085, c == 0x2028, c == 0x2029: //https://en.wikipedia.org/wiki/Newline#Unicode
			// Line Feed, Vertical Tab, Form Feed, Carriage return, Next Line, Line Separator, Paragraph Separator
			return true
		default:
			return false
		}
	}
	// splitting string by space but considering quoted section
	items := strings.FieldsFunc(data, f)
	if s.Data == nil {
		s.Data = make(map[string][]byte, len(items))
	} else {
		//remove all the data from the secret
		for k := range s.Data {
			delete(s.Data, k)
		}
	}
	//Replace secret entries
	encodingSchema, foundEncodingSchema = s.ObjectMeta.Annotations["fortanixEncodingSchema"]
	for _, item := range items {
		x := strings.Split(item, ":")
		//Use unquote to remove escaped quotes from Fortanix output.
		if unquoted, _ = strconv.Unquote(strings.TrimSpace(x[0])); err == nil {
			//if the string is not quoted, unquote returns an empty string
			if unquoted != "" {
				x[0] = unquoted
			}
		}
		if unquoted, err = strconv.Unquote(strings.TrimSpace(x[1])); err == nil {
			if unquoted != "" {
				x[1] = unquoted
			}
		}
		//In Fortanix we store the data in base64
		//During the process of parsing the object to yaml, the data values are base64 encoded
		//because of the json parsing algorithm. If the data is stored in base64 encoding it should
		//be decoded
		if foundEncodingSchema {
			switch encodingSchema {
			case "base64":
				if decoded, err = b64.StdEncoding.DecodeString(strings.TrimSpace(x[1])); err != nil {
					return err
				}
			default:
				decoded = []byte(strings.TrimSpace(x[1]))
			}
		} else {
			//If the annotation is missing, tha data is encoded in base64
			if decoded, err = b64.StdEncoding.DecodeString(strings.TrimSpace(x[1])); err != nil {
				return err
			}
		}
		s.Data[x[0]] = decoded
	}
	// Just in case, clear StringData map
	for k := range s.Data {
		delete(s.StringData, k)
	}

	return nil
}

func (h decryptManifestsCommandHandler) decryptSecuredObjects(yaml string) (string, error) {
	var rcerror, err error
	var secret kms.Secret
	var rObject runtime.Object

	decode := scheme.Codecs.UniversalDeserializer().Decode
	if rObject, _, err = decode([]byte(yaml), nil, nil); err != nil {
		//an unknown CRD will trigger an error decoding the yaml
		return "", errortree.Add(rcerror, "decryptSecuredObjects", err)
	}

	switch o := rObject.(type) {
	case *v1.Secret:
		if _, found := o.ObjectMeta.Annotations["fortanixGroupId"]; found {
			if name, found1 := o.ObjectMeta.Annotations["fortanixSecretName"]; found1 {
				ctx := context.Background()

				if secret, err = h.kmngr.DecryptSecret(ctx, &name); err != nil {
					return "", errortree.Add(rcerror, "decryptSecuredObjects", err)
				}
				if err = setSecretDataFromKeyValueString(o, string(*secret.Blob)); err != nil {
					return "", errortree.Add(rcerror, "decryptSecuredObjects", err)
				}
				//convert object to its yaml representation
				y := printers.YAMLPrinter{}
				builder := strings.Builder{}
				y.PrintObj(o, &builder)

				return builder.String(), nil
			}
		}

		return "", errortree.Add(rcerror, "decryptSecuredObjects", errors.New("removing secret from input"))
	}

	return yaml, nil
}

// Handle Handles the update request
func (h decryptManifestsCommandHandler) Handle(request DecryptManifestsRequest) error {
	var rcerror, err error

	for r := range request.ReceiverCh {
		if r.Yaml, err = h.decryptSecuredObjects(r.Yaml); err != nil {
			//FIXME: decide what to do when there is an error decrypting a secret
			close(request.SendCh)
			return errortree.Add(rcerror, "Handle", err)
		}
		request.SendCh <- r
	} //for
	//Signal no more manifests to process
	close(request.SendCh)

	return nil
}
