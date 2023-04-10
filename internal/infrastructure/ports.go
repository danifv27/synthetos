// The input ports provide the entry points of the application that receive input from the outside world.
//
// For example, an input port could be an HTTP handler handling synchronous calls or a Kafka consumer
// handling asynchronous messages.
package infrastructure

import "github.com/speijnik/go-errortree"

// An PortOption applies optional changes to the Kong application.
type PortOption interface {
	Apply(p *Ports) error
}

// AdapterOptionFunc is function that adheres to the Option interface.
type PortOptionFunc func(p *Ports) error

func (o PortOptionFunc) Apply(p *Ports) error {
	return o(p)
}

// Ports contains the ports services
type Ports struct{}

// NewPorts  instantiates the services of input ports
func NewPorts(opts ...PortOption) (Ports, error) {
	var rcerror error

	p := Ports{}

	// Loop through each option
	for _, option := range opts {
		if err := option.Apply(&p); err != nil {
			return Ports{}, errortree.Add(rcerror, "NewPorts", err)
		}
	}

	return p, nil
}

// PortWithOptions instantiates the services of input ports
func PortWithOptions(p *Ports, opts ...PortOption) error {
	var rcerror error

	// Loop through each option
	for _, option := range opts {
		if err := option.Apply(p); err != nil {
			return errortree.Add(rcerror, "PortWithOptions", err)
		}
	}

	return nil
}
