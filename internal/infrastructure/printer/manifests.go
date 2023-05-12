package printer

import (
	"io"

	"fry.org/cmo/cli/internal/application/provider"
	"github.com/speijnik/go-errortree"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

func (t *PrinterClient) ListManifests(receiveCh <-chan provider.Manifest) error {
	var rcerror error

	e := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)
	for r := range receiveCh {
		io.WriteString(t.wr, "---\n")
		if err := e.Encode(r.Obj, t.wr); err != nil {
			return errortree.Add(rcerror, "ListManifests", err)
		}
	} //for

	return nil
}
