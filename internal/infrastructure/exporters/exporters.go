package exporters

// An ExporterOption applies optional changes to the Kong application.
type ExporterOption interface {
	Apply(t interface{}) error
}

// ExportOptionFn is function that adheres to the ExporterOption interface.
type ExportOptionFn func(t interface{}) error

func (o ExportOptionFn) Apply(t interface{}) error {

	return o(t)
}
