package main

import (
	"context"
	"github.com/Layr-Labs/eigenda/common/configuration"
	"github.com/Layr-Labs/eigenda/lightnode/streaming"
)

// main is the entrypoint for the light node.
func main() {
	config := &streaming.Config{}
	err := configuration.ParseCliArgs(config)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	if config.SourceConfig != nil {
		source := streaming.NewSource(ctx, config)
		err = source.Start()
		if err != nil {
			panic(err)
		}
	} else if config.DestinationConfig != nil {
		dest := streaming.NewDestination(ctx, config)
		err = dest.Start()
		if err != nil {
			panic(err)
		}
	} else {
		panic("either source or destination config must be specified")
	}
}
