package printer

import (
	"io"

	"fry.org/cmo/cli/internal/application/provider"
	"github.com/speijnik/go-errortree"
)

func (t *PrinterClient) ListManifests(receiveCh <-chan provider.Manifest) error {
	var rcerror error

	// e := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)
	for r := range receiveCh {
		io.WriteString(t.wr, "---\n")
		// if err := e.Encode(r.Obj, t.wr); err != nil {
		if _, err := io.WriteString(t.wr, r.Yaml); err != nil {
			return errortree.Add(rcerror, "ListManifests", err)
		}
	} //for

	return nil
}
