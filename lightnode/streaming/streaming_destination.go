package streaming

type Destination struct {
	config *DestinationConfig
}

func NewDestination(config *DestinationConfig) *Destination {
	return &Destination{config: config}
}

func (d *Destination) Start() error {
	return nil
}
