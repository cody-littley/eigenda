package streaming

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/eigenda/api/grpc/lightnode"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
		return d.stream()
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
func (d *Destination) stream() error {

	target := "0.0.0.0:12345"

	conn, err := grpc.Dial(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Printf("failed to close connection: %v\n", err)
		}
	}()

	client := lightnode.NewSourceClient(conn)

	streamClient, err := client.StreamData(d.ctx, &lightnode.StreamDataRequest{})
	if err != nil {
		return nil
	}

	// TODO how to break out if local context is cancelled?
	for {
		reply, err := streamClient.Recv()
		if err != nil {
			return nil
		}
		fmt.Printf("got data of length %s\n", len(reply.Data))
	}
}
