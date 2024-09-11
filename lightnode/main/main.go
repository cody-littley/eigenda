package main

import (
	"github.com/Layr-Labs/eigenda/common/configuration"
	"github.com/Layr-Labs/eigenda/lightnode/streaming"
)

// main is the entrypoint for the light node.
func main() {
	//fmt.Println("Hello world")

	configFilePath := "./data/config.json"
	config := &streaming.Config{}

	err := configuration.ParseJsonFile(config, configFilePath)
	if err != nil {
		panic(err)
	}

}
