package provider

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"fry.org/cmo/cli/internal/application/logger"
	aProvider "fry.org/cmo/cli/internal/application/provider"
	"github.com/speijnik/go-errortree"
)

type yamlReader struct {
	l      logger.Logger
	reader io.Reader
}

// NewReaderProvider creates a new CucumberExporter
func NewReaderProvider(opts ...ProviderOption) (aProvider.ManifestProvider, error) {
	var rcerror error

	c := yamlReader{}
	// Loop through each option
	for _, option := range opts {
		if err := option.Apply(&c); err != nil {
			return nil, errortree.Add(rcerror, "NewKustomizationProvider", err)
		}
	}

	return &c, nil
}

func WithReaderProviderInputType(t string) ProviderOption {

	return ProviderOptionFn(func(i interface{}) error {
		var rcerror, err error
		var m *yamlReader
		var ok bool

		if m, ok = i.(*yamlReader); ok {
			if m.reader, err = getReaderFromInput(t); err != nil {
				return errortree.Add(rcerror, "provider.WithManifestProviderInputType", err)
			}
			return nil
		}

		return errortree.Add(rcerror, "provider.WithManifestProviderInputType", errors.New("type mismatch, yamlReader expected"))
	})
}

func getReaderFromInput(input string) (io.Reader, error) {
	var err, rcerror error
	var fi os.FileInfo

	switch {
	case input == "-":
		if fi, err = os.Stdin.Stat(); err != nil {
			return nil, errortree.Add(rcerror, "getReaderFromInput", err)
		}
		if fi.Mode()&os.ModeNamedPipe == 0 {
			return nil, errortree.Add(rcerror, "getReaderFromInput", errors.New("not a stdin pipe"))
		}
		return bufio.NewReader(os.Stdin), nil
	case strings.Index(input, "http://") == 0 || strings.Index(input, "https://") == 0:
		var resp *http.Response

		_, err = url.Parse(input)
		if err != nil {
			return nil, errortree.Add(rcerror, "getReaderFromInput", err)
		}
		resp, err = http.Get(input)
		if err != nil {
			return nil, errortree.Add(rcerror, "getReaderFromInput", err)
		}
		defer resp.Body.Close()
		return bufio.NewReader(resp.Body), nil
	default:
		var f *os.File

		fi, err = os.Stat(input)
		switch {
		case err != nil:
			return nil, errortree.Add(rcerror, "getReaderFromInput", err)
		case fi.IsDir():
			// it's a dir!
			return nil, errortree.Add(rcerror, "getReaderFromInput", err)
		default:
			// it's a file!
			if f, err = os.Open(input); err != nil {
				return nil, errortree.Add(rcerror, "getReaderFromInput", err)
			}
			return bufio.NewReader(f), nil
		}
	}
}

const yamlSeparator = "\n---"

// const separator = "---"

// splitYAMLDocument is a bufio.SplitFunc for splitting YAML streams into individual documents.
func splitYAMLDocument(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	sep := len([]byte(yamlSeparator))
	if i := bytes.Index(data, []byte(yamlSeparator)); i >= 0 {
		// We have a potential document terminator
		i += sep
		after := data[i:]
		if len(after) == 0 {
			// we can't read any more characters
			if atEOF {
				return len(data), data[:len(data)-sep], nil
			}
			return 0, nil, nil
		}
		if j := bytes.IndexByte(after, '\n'); j >= 0 {
			return i + j + 1, data[0 : i-sep], nil
		}
		return 0, nil, nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

func (r yamlReader) GetManifests(ctx context.Context, sendCh chan<- aProvider.Manifest) error {
	var rcerror, err error
	var scanner *bufio.Scanner
	var m aProvider.Manifest

	defer close(sendCh)
	scanner = bufio.NewScanner(r.reader)
	buf := make([]byte, 0, 64*1024)
	//Sets the maximum token size.
	//We will be able to scan the file as long as none of the lines is larger than 1MB.
	scanner.Buffer(buf, 1024*1024)
	scanner.Split(splitYAMLDocument)
	for scanner.Scan() {
		// decode := scheme.Codecs.UniversalDeserializer().Decode
		o := scanner.Bytes()
		// if m.Obj, _, err = decode(o, nil, nil); err != nil {
		// 	//an unknown CRD will trigger an error decoding the yaml
		// 	r.l.WithFields(logger.Fields{
		// 		"err": err,
		// 	}).Debug("unable to decode yaml object")
		// 	continue
		// }
		m.Yaml = string(o)
		sendCh <- m
	}
	//After we exit the loop, we have have either hit the end of the input, or encountered an error.
	if err = scanner.Err(); err != nil {
		return errortree.Add(rcerror, "GetManifests", err)
	}

	return nil
}