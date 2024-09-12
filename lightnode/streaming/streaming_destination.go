package streaming

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/eigenda/api/grpc/lightnode"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"sync"
)

type Destination struct {
	ctx    context.Context
	config *DestinationConfig
}

func NewDestination(
	ctx context.Context,
	config *DestinationConfig) *Destination {
	if config == nil {
		panic("config cannot be nil")
	}
	return &Destination{
		ctx:    ctx,
		config: config,
	}
}

func (d *Destination) Start() error {
	switch d.config.TransferStrategy {
	case Stream:
		d.stream()
	case Put:
		// No need to do anything, source will initiate
		select {
		case <-d.ctx.Done():
			fmt.Println("context cancelled, exiting")
		}
	case Get:
		// TODO implement me
		panic("implement me")
	}

	return nil
}

// stream streams data from the source to the destination
func (d *Destination) stream() {

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(d.config.NumberOfConnections)

	for i := 0; i < d.config.NumberOfConnections; i++ {
		go func() {
			defer waitGroup.Done()

			target := d.config.SourceHostname

			fmt.Printf("dialing %s\n", target)

			conn, err := grpc.DialContext(
				d.ctx,
				d.config.SourceHostname,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithBlock())
			if err != nil {
				fmt.Printf("failed to dial: %v\n", err)
			}
			defer conn.Close()

			fmt.Println("connection established")

			client := lightnode.NewSourceClient(conn)

			streamClient, err := client.StreamData(d.ctx, &lightnode.StreamDataRequest{})
			if err != nil {
				fmt.Printf("failed to stream data: %v\n", err)
			}

			fmt.Println("about to enter loop") // TODO

			// TODO how to break out if local context is cancelled?
			for {
				_, err := streamClient.Recv()

				if err == io.EOF {
					break
				} else if err != nil {
					fmt.Printf("failed to receive: %v\n", err)
					break
				}
			}
		}()
	}

	waitGroup.Wait()
}
