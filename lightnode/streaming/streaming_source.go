package streaming

type Source struct {
	config *SourceConfig
}

func NewSource(config *SourceConfig) *Source {
	return &Source{config: config}
}

func (s *Source) Start() error {
	return nil
}
