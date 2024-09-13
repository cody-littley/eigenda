package streaming

import (
	"context"
	"fmt"
	"github.com/Layr-Labs/eigenda/api/grpc/lightnode"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"net"
	"sync"
)

type Destination struct {
	lightnode.DestinationServer
	ctx    context.Context
	config *Config
}

func NewDestination(
	ctx context.Context,
	config *Config) *Destination {
	if config == nil {
		panic("config cannot be nil")
	}
	return &Destination{
		ctx:    ctx,
		config: config,
	}
}

func (d *Destination) Start() error {
	switch d.config.DestinationConfig.TransferStrategy {
	case Stream:
		d.stream()
	case Put:
		return d.put()
	}
	return nil
}

// stream streams data from the source to the destination
func (d *Destination) stream() {

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(d.config.DestinationConfig.NumberOfConnections)

	for i := 0; i < d.config.DestinationConfig.NumberOfConnections; i++ {
		go func() {
			defer waitGroup.Done()

			target := d.config.DestinationConfig.SourceHostname

			fmt.Printf("dialing %s\n", target)

			conn, err := grpc.DialContext(
				d.ctx,
				d.config.DestinationConfig.SourceHostname,
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

func (d *Destination) put() error {

	fmt.Printf("listening on :%d\n", d.config.Port)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", d.config.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	fmt.Println("Creating GRPC server")

	grpcServer := grpc.NewServer()

	lightnode.RegisterDestinationServer(grpcServer, d)

	fmt.Println("attaching GRPC server to socket")

	go func() {
		err = grpcServer.Serve(listener)
		if err != nil {
			panic(err)
		}
	}()

	fmt.Println("Server online")

	////////////////////////////////////

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(d.config.DestinationConfig.NumberOfConnections)

	for i := 0; i < d.config.DestinationConfig.NumberOfConnections; i++ {
		target := d.config.DestinationConfig.SourceHostname

		fmt.Printf("dialing %s\n", target)

		go func() {
			conn, err := grpc.DialContext(
				d.ctx,
				d.config.DestinationConfig.SourceHostname,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithBlock())
			if err != nil {
				fmt.Printf("failed to dial: %v\n", err)
			}
			defer conn.Close()

			fmt.Println("connection established")

			client := lightnode.NewSourceClient(conn)

			selfAddress := fmt.Sprintf("%s:%d", d.config.DestinationConfig.SelfIP, d.config.Port)

			_, err = client.RequestPushes(d.ctx, &lightnode.RequestPushesRequest{
				Destination: selfAddress,
			})
			if err != nil {
				fmt.Printf("failed to request pushes: %v\n", err)
			} else {
				fmt.Println("pushes requested")
			}
			waitGroup.Done()
		}()
	}

	waitGroup.Wait()

	select {
	case <-d.ctx.Done():
		fmt.Println("context cancelled, exiting")
		grpcServer.GracefulStop()
	}

	fmt.Println("Server shutting down")
	return nil
}

func (d *Destination) ReceiveData(ctx context.Context, data *lightnode.ReceiveDataRequest) (*lightnode.ReceiveDataReply, error) {
	return &lightnode.ReceiveDataReply{}, nil
}
