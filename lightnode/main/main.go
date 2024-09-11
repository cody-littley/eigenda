package main

import (
	"context"
	"github.com/Layr-Labs/eigenda/common/configuration"
	"github.com/Layr-Labs/eigenda/lightnode/streaming"
)

// main is the entrypoint for the light node.
func main() {

	configFilePath := "./data/config.json"
	config := &streaming.Config{}

	err := configuration.ParseJsonFile(config, configFilePath)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	if config.SourceConfig != nil {
		source := streaming.NewSource(ctx, config.SourceConfig)
		err := source.Start()
		if err != nil {
			panic(err)
		}
	} else {
		dest := streaming.NewDestination(ctx, config.DestinationConfig)
		err := dest.Start()
		if err != nil {
			panic(err)
		}
	}
}
